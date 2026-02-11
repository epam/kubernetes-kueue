/*
Copyright The Kubernetes Authors.

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

package controller

import (
	"context"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kueue "sigs.k8s.io/kueue/apis/kueue/v1beta2"
	"sigs.k8s.io/kueue/cmd/experimental/priority-boost-controller/pkg/constants"
)

// PriorityBoostReconciler watches Workload objects and sets the
// kueue.x-k8s.io/priority-boost annotation to protect admitted workloads
// from same-priority preemption for at least minAdmitDuration after admission.
type PriorityBoostReconciler struct {
	client           client.Client
	log              logr.Logger
	recorder         record.EventRecorder
	minAdmitDuration time.Duration
	boostValue       int32
}

var _ reconcile.Reconciler = (*PriorityBoostReconciler)(nil)

type PriorityBoostReconcilerOptions struct {
	MinAdmitDuration time.Duration
	BoostValue       int32
}

type PriorityBoostReconcilerOption func(*PriorityBoostReconcilerOptions)

func WithMinAdmitDuration(d time.Duration) PriorityBoostReconcilerOption {
	return func(o *PriorityBoostReconcilerOptions) {
		o.MinAdmitDuration = d
	}
}

func WithBoostValue(v int32) PriorityBoostReconcilerOption {
	return func(o *PriorityBoostReconcilerOptions) {
		o.BoostValue = v
	}
}

var defaultOptions = PriorityBoostReconcilerOptions{
	MinAdmitDuration: 0,
	BoostValue:       100000,
}

func NewPriorityBoostReconciler(c client.Client, recorder record.EventRecorder, opts ...PriorityBoostReconcilerOption) *PriorityBoostReconciler {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}
	return &PriorityBoostReconciler{
		client:           c,
		log:              ctrl.Log.WithName(constants.ControllerName),
		recorder:         recorder,
		minAdmitDuration: options.MinAdmitDuration,
		boostValue:       options.BoostValue,
	}
}

// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=kueue.x-k8s.io,resources=workloads,verbs=get;list;watch;patch
func (r *PriorityBoostReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	var wl kueue.Workload
	if err := r.client.Get(ctx, req.NamespacedName, &wl); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if !wl.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	newBoost, requeueAfter := r.computeBoost(&wl)
	currentBoostValue := wl.Annotations[constants.PriorityBoostAnnotationKey]

	var newBoostValue string
	if newBoost > 0 {
		newBoostValue = strconv.FormatInt(int64(newBoost), 10)
	}
	// newBoostValue == "" means the annotation should be absent.

	if currentBoostValue == newBoostValue {
		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}

	patch := client.MergeFrom(wl.DeepCopy())
	if newBoost > 0 {
		if wl.Annotations == nil {
			wl.Annotations = make(map[string]string)
		}
		wl.Annotations[constants.PriorityBoostAnnotationKey] = newBoostValue
		log.Info("Setting priority-boost annotation (workload within minAdmitDuration window)",
			"workload", klog.KObj(&wl),
			"annotationKey", constants.PriorityBoostAnnotationKey,
			"annotationValue", newBoostValue,
			"requeueAfter", requeueAfter)
		r.recorder.Eventf(&wl, corev1.EventTypeNormal, "PriorityBoostSet",
			"Set %s=%s (minAdmitDuration protection window active)",
			constants.PriorityBoostAnnotationKey, newBoostValue)
	} else {
		delete(wl.Annotations, constants.PriorityBoostAnnotationKey)
		log.Info("Removing priority-boost annotation (minAdmitDuration window expired or workload not admitted)",
			"workload", klog.KObj(&wl),
			"annotationKey", constants.PriorityBoostAnnotationKey)
		r.recorder.Eventf(&wl, corev1.EventTypeNormal, "PriorityBoostCleared",
			"Cleared %s (minAdmitDuration protection window expired)",
			constants.PriorityBoostAnnotationKey)
	}

	if err := r.client.Patch(ctx, &wl, patch); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

func (r *PriorityBoostReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return builder.TypedControllerManagedBy[reconcile.Request](mgr).
		For(&kueue.Workload{}).
		Complete(r)
}

// computeBoost returns the boost value to apply to the workload and how long
// until the controller should re-evaluate.
//
// When minAdmitDuration is configured and the workload has been admitted for
// less than that duration, the boost is set to boostValue so that same-priority
// pending workloads cannot preempt this workload. Once the window expires, the
// boost is removed (0) and requeueAfter is 0.
func (r *PriorityBoostReconciler) computeBoost(wl *kueue.Workload) (boost int32, requeueAfter time.Duration) {
	if r.minAdmitDuration <= 0 {
		return 0, 0
	}

	admittedCond := meta.FindStatusCondition(wl.Status.Conditions, string(kueue.WorkloadAdmitted))
	if admittedCond == nil || admittedCond.Status != metav1.ConditionTrue {
		// Workload is not currently admitted — no protection needed.
		return 0, 0
	}

	elapsed := time.Since(admittedCond.LastTransitionTime.Time)
	if elapsed < r.minAdmitDuration {
		remaining := r.minAdmitDuration - elapsed
		return r.boostValue, remaining
	}

	// Protection window has expired.
	return 0, 0
}
