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
	"maps"
	"slices"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	utilstatefulset "sigs.k8s.io/kueue/pkg/util/statefulset"
	"sigs.k8s.io/kueue/pkg/workload"
)

const (
	podBatchPeriod   = time.Second
	leaderPodSetName = "leader"
	workerPodSetName = "worker"
)

type Reconciler struct {
	client          client.Client
	record          record.EventRecorder
	labelKeysToCopy []string
}

func NewReconciler(client client.Client, eventRecorder record.EventRecorder, opts ...jobframework.Option) jobframework.JobReconcilerInterface {
	options := jobframework.ProcessOptions(opts...)

	return &Reconciler{
		client:          client,
		record:          eventRecorder,
		labelKeysToCopy: options.LabelKeysToCopy,
	}
}

var _ jobframework.JobReconcilerInterface = (*Reconciler)(nil)

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctrl.Log.V(3).Info("Setting up LeaderWorkerSet reconciler")

	return ctrl.NewControllerManagedBy(mgr).
		For(&leaderworkersetv1.LeaderWorkerSet{}).
		Named("leaderworkerset").
		WithEventFilter(r).
		Watches(&corev1.Pod{}, &podHandler{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=leaderworkerset.x-k8s.io,resources=leaderworkersets,verbs=get;list;watch

func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := ctrl.LoggerFrom(ctx).WithValues("leaderworkerset", klog.KRef(req.Namespace, req.Name))
	ctx = ctrl.LoggerInto(ctx, log)
	log.V(2).Info("Reconciling LeaderWorkerSet")

	pods := &corev1.PodList{}
	if err := r.client.List(ctx, pods, client.InNamespace(req.Namespace), client.MatchingLabels{
		leaderworkersetv1.SetNameLabelKey: req.Name,
	}); err != nil {
		return ctrl.Result{}, err
	}

	lws, err := getLeaderWorkerSet(ctx, r.client, req.NamespacedName)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.createPrebuildWorkloadsIfNotExist(ctx, lws)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.syncPods(ctx, req.Namespace, lws, pods.Items)

	return ctrl.Result{}, err
}

func (r *Reconciler) createPrebuildWorkloadsIfNotExist(ctx context.Context, lws *leaderworkersetv1.LeaderWorkerSet) error {
	if lws == nil {
		return nil
	}
	replicas := ptr.Deref(lws.Spec.Replicas, 1)
	for i := int32(0); i < replicas; i++ {
		if err := r.createPrebuildWorkloadIfNotExist(ctx, lws, i); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) createPrebuildWorkloadIfNotExist(ctx context.Context, lws *leaderworkersetv1.LeaderWorkerSet, index int32) error {
	wl := &kueue.Workload{}
	err := r.client.Get(ctx, client.ObjectKey{Name: GetWorkloadName(lws.UID, lws.Name, fmt.Sprint(index)), Namespace: lws.Namespace}, wl)
	// Ignore if the Workload already exists or an error occurs.
	if err == nil || client.IgnoreNotFound(err) != nil {
		return err
	}
	return r.createPrebuildWorkload(ctx, lws, index)
}

func (r *Reconciler) createPrebuildWorkload(ctx context.Context, lws *leaderworkersetv1.LeaderWorkerSet, index int32) error {
	createdWorkload := r.constructWorkload(lws, index)
	err := r.client.Create(ctx, createdWorkload)
	if err != nil {
		return err
	}
	r.record.Eventf(
		lws, corev1.EventTypeNormal, jobframework.ReasonCreatedWorkload,
		"Created Workload: %v", workload.Key(createdWorkload),
	)
	return nil
}

func (r *Reconciler) constructWorkload(lws *leaderworkersetv1.LeaderWorkerSet, index int32) *kueue.Workload {
	return podcontroller.NewGroupWorkload(GetWorkloadName(lws.UID, lws.Name, fmt.Sprint(index)), lws, r.podSets(lws), r.labelKeysToCopy)
}

func (r *Reconciler) podSets(lws *leaderworkersetv1.LeaderWorkerSet) []kueue.PodSet {
	podSets := make([]kueue.PodSet, 0, 2)

	if lws.Spec.LeaderWorkerTemplate.LeaderTemplate != nil {
		podSets = append(podSets, kueue.PodSet{
			Name:  leaderPodSetName,
			Count: 1,
			Template: corev1.PodTemplateSpec{
				Spec: *lws.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.DeepCopy(),
			},
			TopologyRequest: jobframework.PodSetTopologyRequest(
				&lws.Spec.LeaderWorkerTemplate.LeaderTemplate.ObjectMeta,
				nil,
				nil,
				nil,
			),
		})
	}

	defaultPodSetName := kueue.DefaultPodSetName
	if len(podSets) > 0 {
		defaultPodSetName = workerPodSetName
	}

	defaultPodSetCount := ptr.Deref(lws.Spec.LeaderWorkerTemplate.Size, 1)
	if len(podSets) > 0 {
		defaultPodSetCount--
	}

	podSets = append(podSets, kueue.PodSet{
		Name:  defaultPodSetName,
		Count: defaultPodSetCount,
		Template: corev1.PodTemplateSpec{
			Spec: *lws.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.DeepCopy(),
		},
		TopologyRequest: jobframework.PodSetTopologyRequest(
			&lws.Spec.LeaderWorkerTemplate.WorkerTemplate.ObjectMeta,
			nil,
			nil,
			nil,
		),
	})

	return podSets
}

func (r *Reconciler) syncPods(ctx context.Context, namespace string, lws *leaderworkersetv1.LeaderWorkerSet, pods []corev1.Pod) error {
	stsPods := make(map[string][]*corev1.Pod)
	for _, pod := range pods {
		var stsName string
		if controllerRef := metav1.GetControllerOf(&pod); controllerRef != nil {
			stsName = controllerRef.Name
		}
		stsPods[stsName] = append(stsPods[stsName], &pod)
	}

	stsNames := slices.Collect(maps.Keys(stsPods))
	return parallelize.Until(ctx, len(stsNames), func(i int) error {
		stsName := stsNames[i]
		return r.syncPodGroup(ctx, namespace, lws, stsName, stsPods[stsName])
	})
}

func (r *Reconciler) syncPodGroup(ctx context.Context, namespace string, lws *leaderworkersetv1.LeaderWorkerSet, stsName string, pods []*corev1.Pod) error {
	sts, err := utilstatefulset.GetStatefulSet(ctx, r.client, client.ObjectKey{Name: stsName, Namespace: namespace})
	if err != nil {
		return err
	}
	return parallelize.Until(ctx, len(pods), func(i int) error {
		return r.syncPod(ctx, lws, sts, pods[i])
	})
}

func (r *Reconciler) syncPod(ctx context.Context, lws *leaderworkersetv1.LeaderWorkerSet, sts *appsv1.StatefulSet, pod *corev1.Pod) error {
	log := ctrl.LoggerFrom(ctx)
	return client.IgnoreNotFound(clientutil.Patch(ctx, r.client, pod, true, func() (bool, error) {
		if utilstatefulset.UngateAndFinalizePod(sts, pod) {
			log.V(3).Info(
				"Finalizing pod in group",
				"pod", klog.KObj(pod),
				"group", pod.Labels[podcontroller.GroupNameLabel],
			)
			return true, nil
		}

		if r.setDefault(lws, pod) {
			log.V(3).Info(
				"Updating pod in group", "pod",
				klog.KObj(pod), "group",
				pod.Labels[podcontroller.GroupNameLabel],
			)
			return true, nil
		}

		return false, nil
	}))
}

func (r *Reconciler) setDefault(lws *leaderworkersetv1.LeaderWorkerSet, pod *corev1.Pod) bool {
	if lws == nil {
		return false
	}

	// If queue label already exist nothing to update.
	if _, ok := pod.Labels[constants.QueueLabel]; ok {
		return false
	}

	// We should wait for GroupIndexLabelKey.
	if _, ok := pod.Labels[leaderworkersetv1.GroupIndexLabelKey]; !ok {
		return false
	}

	queueName := jobframework.QueueNameForObject(lws)
	// Ignore LeaderWorkerSet without queue name.
	if queueName == "" {
		return false
	}

	wlName := GetWorkloadName(lws.UID, lws.Name, pod.Labels[leaderworkersetv1.GroupIndexLabelKey])

	pod.Labels[constants.QueueLabel] = queueName
	pod.Labels[podcontroller.GroupNameLabel] = wlName
	pod.Labels[constants.PrebuiltWorkloadLabel] = wlName
	if priorityClass := jobframework.WorkloadPriorityClassName(lws); priorityClass != "" {
		pod.Labels[constants.WorkloadPriorityClassLabel] = priorityClass
	}

	pod.Annotations[podcontroller.GroupTotalCountAnnotation] = fmt.Sprint(ptr.Deref(lws.Spec.LeaderWorkerTemplate.Size, 1))
	pod.Annotations[podcontroller.RoleHashAnnotation] = kueue.DefaultPodSetName

	if lws.Spec.LeaderWorkerTemplate.LeaderTemplate != nil {
		if _, ok := pod.Annotations[leaderworkersetv1.LeaderPodNameAnnotationKey]; !ok {
			pod.Annotations[podcontroller.RoleHashAnnotation] = leaderPodSetName
		} else {
			pod.Annotations[podcontroller.RoleHashAnnotation] = workerPodSetName
		}
	}

	return true
}

var _ predicate.Predicate = (*Reconciler)(nil)

func (r *Reconciler) Generic(event.GenericEvent) bool {
	return false
}

func (r *Reconciler) Create(e event.CreateEvent) bool {
	return r.handle(e.Object)
}

func (r *Reconciler) Update(e event.UpdateEvent) bool {
	return r.handle(e.ObjectNew)
}

func (r *Reconciler) Delete(e event.DeleteEvent) bool {
	return r.handle(e.Object)
}

func (r *Reconciler) handle(obj client.Object) bool {
	lws, isLws := obj.(*leaderworkersetv1.LeaderWorkerSet)
	if !isLws {
		return true
	}
	// Handle only leaderworkerset managed by kueue.
	return jobframework.IsManagedByKueue(lws)
}

var _ handler.EventHandler = (*podHandler)(nil)

type podHandler struct{}

func (h *podHandler) Generic(context.Context, event.GenericEvent, workqueue.TypedRateLimitingInterface[reconcile.Request]) {
}

func (h *podHandler) Create(_ context.Context, e event.CreateEvent, q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	// To watch for pod ADDED events.
	h.handle(e.Object, q)
}

func (h *podHandler) Update(_ context.Context, e event.UpdateEvent, q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
}

func (h *podHandler) Delete(context.Context, event.DeleteEvent, workqueue.TypedRateLimitingInterface[reconcile.Request]) {
}

func (h *podHandler) handle(obj client.Object, q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	pod, isPod := obj.(*corev1.Pod)
	if !isPod || pod.Annotations[podcontroller.SuspendedByParentAnnotation] != FrameworkName {
		return
	}
	if controllerRef := metav1.GetControllerOf(pod); controllerRef != nil {
		if controllerRef.Kind != "StatefulSet" || controllerRef.APIVersion != "apps/v1" {
			// The pod is controlled by an owner that is not an apps/v1 StatefulSet.
			return
		}
		q.AddAfter(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: pod.Namespace,
				Name:      controllerRef.Name,
			},
		}, podBatchPeriod)
	}
}

func getLeaderWorkerSet(ctx context.Context, c client.Client, key types.NamespacedName) (*leaderworkersetv1.LeaderWorkerSet, error) {
	if key.Name == "" || key.Namespace == "" {
		return nil, nil
	}
	lws := &leaderworkersetv1.LeaderWorkerSet{}
	if err := c.Get(ctx, key, lws); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return lws, nil
}
