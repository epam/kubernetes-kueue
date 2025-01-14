/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package leaderworkerset

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	leaderworkersetv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"

	kueue "sigs.k8s.io/kueue/apis/kueue/v1beta1"
	"sigs.k8s.io/kueue/pkg/controller/constants"
	"sigs.k8s.io/kueue/pkg/controller/jobframework"
	podcontroller "sigs.k8s.io/kueue/pkg/controller/jobs/pod"
	clientutil "sigs.k8s.io/kueue/pkg/util/client"
	"sigs.k8s.io/kueue/pkg/util/parallelize"
	utilpod "sigs.k8s.io/kueue/pkg/util/pod"
)

const (
	batchPeriod = time.Second
)

type Reconciler struct {
	client     client.Client
	podHandler *podHandler
}

func NewReconciler(client client.Client, _ record.EventRecorder, _ ...jobframework.Option) jobframework.JobReconcilerInterface {
	return &Reconciler{
		client:     client,
		podHandler: &podHandler{},
	}
}

var _ jobframework.JobReconcilerInterface = (*Reconciler)(nil)

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctrl.Log.V(3).Info("Setting up LeaderWorkerSet reconciler")
	return ctrl.NewControllerManagedBy(mgr).
		For(&leaderworkersetv1.LeaderWorkerSet{}).
		Named("leaderworkerset").
		WithOptions(controller.Options{NeedLeaderElection: ptr.To(false)}).
		Watches(&corev1.Pod{}, r.podHandler).
		WithEventFilter(r).
		Complete(r)
}

// +kubebuilder:rbac:groups=leaderworkerset.x-k8s.io,resources=leaderworkersets,verbs=get;list;watch;patch
// +kubebuilder:rbac:groups=leaderworkerset.x-k8s.io,resources=leaderworkersets/finalizers,verbs=patch

func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	lws := &leaderworkersetv1.LeaderWorkerSet{}
	err := r.client.Get(ctx, req.NamespacedName, lws)
	if err != nil {
		// we'll ignore not-found errors, since there is nothing to do.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log := ctrl.LoggerFrom(ctx).WithValues("leaderworkerset", klog.KObj(lws))
	ctx = ctrl.LoggerInto(ctx, log)
	log.V(2).Info("Reconciling LeaderWorkerSet")

	if lws.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(lws, kueue.ResourceInUseFinalizerName) {
		err := client.IgnoreNotFound(clientutil.Patch(ctx, r.client, lws, true, func() (bool, error) {
			controllerutil.AddFinalizer(lws, kueue.ResourceInUseFinalizerName)
			log.V(5).Info("Added finalizer")
			return true, nil
		}))
		return ctrl.Result{}, err
	}

	podList := &corev1.PodList{}
	if err := r.client.List(ctx, podList, client.InNamespace(lws.Namespace), client.MatchingLabels{
		leaderworkersetv1.SetNameLabelKey: lws.Name,
	}); err != nil {
		return ctrl.Result{}, err
	}

	if err = parallelize.Until(ctx, len(podList.Items), func(i int) error {
		pod := &podList.Items[i]
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed || !lws.DeletionTimestamp.IsZero() {
			return client.IgnoreNotFound(clientutil.Patch(ctx, r.client, pod, true, func() (bool, error) {
				removed := controllerutil.RemoveFinalizer(pod, podcontroller.PodFinalizer)
				if removed {
					log.V(3).Info(
						"Finalizing leaderworkerset pod in the group",
						"leaderworkerset", lws.Name,
						"pod", klog.KObj(pod),
						"group", pod.Labels[podcontroller.GroupNameLabel],
					)
				}
				return removed, nil
			}))
		} else {
			return client.IgnoreNotFound(clientutil.Patch(ctx, r.client, pod, true, func() (bool, error) {
				updated, err := r.setDefault(lws, pod)
				if err != nil {
					return false, err
				}
				if updated {
					log.V(3).Info(
						"Set default values for leaderworkerset pod in the group",
						"leaderworkerset", lws.Name,
						"pod", klog.KObj(pod),
						"group", pod.Labels[podcontroller.GroupNameLabel],
					)
				}
				return updated, nil
			}))
		}
	}); err != nil {
		return ctrl.Result{}, err
	}

	// We should wait for all pods to be finalized, and only after that we can delete the LeaderWorkerSet.
	if !lws.DeletionTimestamp.IsZero() && controllerutil.ContainsFinalizer(lws, kueue.ResourceInUseFinalizerName) {
		err := client.IgnoreNotFound(clientutil.Patch(ctx, r.client, lws, true, func() (bool, error) {
			controllerutil.RemoveFinalizer(lws, kueue.ResourceInUseFinalizerName)
			log.V(5).Info("Removed finalizer")
			return true, nil
		}))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) setDefault(lws *leaderworkersetv1.LeaderWorkerSet, pod *corev1.Pod) (bool, error) {
	// If queue label already exist nothing to update.
	if _, ok := pod.Labels[constants.QueueLabel]; ok {
		return false, nil
	}

	// We should wait for GroupIndexLabelKey.
	if _, ok := pod.Labels[leaderworkersetv1.GroupIndexLabelKey]; !ok {
		return false, nil
	}

	groupName, err := GetWorkloadName(lws, pod.Labels[leaderworkersetv1.GroupIndexLabelKey])
	if err != nil {
		return false, err
	}

	pod.Labels[constants.QueueLabel] = jobframework.QueueNameForObject(lws)
	pod.Labels[podcontroller.GroupNameLabel] = groupName
	pod.Annotations[podcontroller.GroupTotalCountAnnotation] = fmt.Sprint(ptr.Deref(lws.Spec.LeaderWorkerTemplate.Size, 1))

	hash, err := utilpod.GenerateShape(pod.Spec)
	if err != nil {
		return false, err
	}
	pod.Annotations[podcontroller.RoleHashAnnotation] = hash

	return true, nil
}

var _ predicate.Predicate = (*Reconciler)(nil)

func (r *Reconciler) Generic(event.GenericEvent) bool {
	return false
}

func (r *Reconciler) Create(e event.CreateEvent) bool {
	return r.filter(e.Object)
}

func (r *Reconciler) Update(e event.UpdateEvent) bool {
	return r.filter(e.ObjectNew)
}

func (r *Reconciler) Delete(e event.DeleteEvent) bool {
	_, isLWS := e.Object.(*leaderworkersetv1.LeaderWorkerSet)
	return !isLWS
}

func (r *Reconciler) filter(obj client.Object) bool {
	lws, isLWS := obj.(*leaderworkersetv1.LeaderWorkerSet)
	if !isLWS {
		return true
	}
	// Reconcile only leaderworkerset with queue-name.
	return jobframework.QueueNameForObject(lws) != ""
}

type podHandler struct{}

var _ handler.EventHandler = (*podHandler)(nil)

func (h *podHandler) Generic(context.Context, event.GenericEvent, workqueue.TypedRateLimitingInterface[reconcile.Request]) {
}

func (h *podHandler) Create(_ context.Context, e event.CreateEvent, q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	h.handle(e.Object, q)
}

func (h *podHandler) Update(_ context.Context, e event.UpdateEvent, q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	h.handle(e.ObjectNew, q)
}

func (h *podHandler) Delete(_ context.Context, e event.DeleteEvent, q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	h.handle(e.Object, q)
}

func (h *podHandler) handle(obj client.Object, q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	pod, isPod := obj.(*corev1.Pod)
	if !isPod {
		return
	}
	// Handle only leaderworkerset pods
	if pod.Annotations[podcontroller.SuspendedByParentAnnotation] != FrameworkName {
		return
	}
	q.AddAfter(reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      pod.Labels[leaderworkersetv1.SetNameLabelKey],
			Namespace: pod.Namespace,
		},
	}, batchPeriod)
}
