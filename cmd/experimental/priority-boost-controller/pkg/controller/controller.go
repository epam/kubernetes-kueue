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
	"k8s.io/apimachinery/pkg/labels"
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
// kueue.x-k8s.io/priority-boost annotation after minAdmitDuration as a
// post-window preemption signal (negative boost) when withinClusterQueue
// preemption is LowerPriority. No annotation is set during the window.
type PriorityBoostReconciler struct {
	client              client.Client
	log                 logr.Logger
	recorder            record.EventRecorder
	minAdmitDuration    time.Duration
	boostValue          int32
	workloadSelector    labels.Selector
	maxWorkloadPriority *int32
}

var _ reconcile.Reconciler = (*PriorityBoostReconciler)(nil)

type PriorityBoostReconcilerOptions struct {
	MinAdmitDuration    time.Duration
	BoostValue          int32
	WorkloadSelector    labels.Selector
	MaxWorkloadPriority *int32
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

func WithWorkloadSelector(s labels.Selector) PriorityBoostReconcilerOption {
	return func(o *PriorityBoostReconcilerOptions) {
		o.WorkloadSelector = s
	}
}

func WithMaxWorkloadPriority(p *int32) PriorityBoostReconcilerOption {
	return func(o *PriorityBoostReconcilerOptions) {
		o.MaxWorkloadPriority = p
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
		client:              c,
		log:                 ctrl.Log.WithName(constants.ControllerName),
		recorder:            recorder,
		minAdmitDuration:    options.MinAdmitDuration,
		boostValue:          options.BoostValue,
		workloadSelector:    options.WorkloadSelector,
		maxWorkloadPriority: options.MaxWorkloadPriority,
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

	var newBoost int32
	var requeueAfter time.Duration
	if !r.workloadInScope(&wl) {
		newBoost, requeueAfter = 0, 0
	} else {
		newBoost, requeueAfter = r.computeBoost(&wl)
	}

	currentBoostValue := wl.Annotations[constants.PriorityBoostAnnotationKey]

	var desiredBoostStr string
	if newBoost != 0 {
		desiredBoostStr = strconv.FormatInt(int64(newBoost), 10)
	}

	if currentBoostValue == desiredBoostStr {
		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}

	patch := client.MergeFrom(wl.DeepCopy())
	if newBoost != 0 {
		if wl.Annotations == nil {
			wl.Annotations = make(map[string]string)
		}
		wl.Annotations[constants.PriorityBoostAnnotationKey] = desiredBoostStr
		log.Info("Setting priority-boost annotation (post-window preemption signal)",
			"workload", klog.KObj(&wl),
			"annotationKey", constants.PriorityBoostAnnotationKey,
			"annotationValue", desiredBoostStr,
			"requeueAfter", requeueAfter)
		r.recorder.Eventf(&wl, corev1.EventTypeNormal, "PriorityBoostSet",
			"Set %s=%s (post-window preemption signal)",
			constants.PriorityBoostAnnotationKey, desiredBoostStr)
	} else {
		delete(wl.Annotations, constants.PriorityBoostAnnotationKey)
		log.Info("Removing priority-boost annotation",
			"workload", klog.KObj(&wl),
			"annotationKey", constants.PriorityBoostAnnotationKey)
		r.recorder.Eventf(&wl, corev1.EventTypeNormal, "PriorityBoostCleared",
			"Cleared %s (workload not in scope, not admitted, or feature disabled)",
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

func (r *PriorityBoostReconciler) workloadInScope(wl *kueue.Workload) bool {
	if r.workloadSelector != nil && !r.workloadSelector.Matches(labels.Set(wl.Labels)) {
		return false
	}
	if r.maxWorkloadPriority != nil {
		var p int32
		if wl.Spec.Priority != nil {
			p = *wl.Spec.Priority
		}
		if p > *r.maxWorkloadPriority {
			return false
		}
	}
	return true
}

// computeBoost returns the signed boost to apply and how long until the
// controller should re-evaluate.
//
// With minAdmitDuration > 0: while admitted and elapsed < minAdmitDuration,
// returns (0, remaining). While admitted after the window, returns
// (-boostValue, 0). Otherwise returns (0, 0).
func (r *PriorityBoostReconciler) computeBoost(wl *kueue.Workload) (boost int32, requeueAfter time.Duration) {
	if r.minAdmitDuration <= 0 {
		return 0, 0
	}

	admittedCond := meta.FindStatusCondition(wl.Status.Conditions, string(kueue.WorkloadAdmitted))
	if admittedCond == nil || admittedCond.Status != metav1.ConditionTrue {
		return 0, 0
	}

	elapsed := time.Since(admittedCond.LastTransitionTime.Time)
	if elapsed < r.minAdmitDuration {
		remaining := r.minAdmitDuration - elapsed
		return 0, remaining
	}

	return -r.boostValue, 0
}
