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
	"testing"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	kueue "sigs.k8s.io/kueue/apis/kueue/v1beta2"
	"sigs.k8s.io/kueue/cmd/experimental/priority-boost-controller/pkg/constants"
)

func admittedWorkload(admittedAt time.Time) *kueue.Workload {
	return &kueue.Workload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wl-1",
			Namespace: "ns",
		},
		Status: kueue.WorkloadStatus{
			Conditions: []metav1.Condition{
				{
					Type:               string(kueue.WorkloadAdmitted),
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(admittedAt),
					Reason:             "Admitted",
				},
			},
		},
	}
}

func TestComputeBoost_NotAdmitted(t *testing.T) {
	r := &PriorityBoostReconciler{
		minAdmitDuration: 30 * time.Minute,
		boostValue:       100000,
	}
	wl := &kueue.Workload{}
	boost, requeueAfter := r.computeBoost(wl)
	if boost != 0 {
		t.Errorf("expected boost=0, got %d", boost)
	}
	if requeueAfter != 0 {
		t.Errorf("expected requeueAfter=0, got %v", requeueAfter)
	}
}

func TestComputeBoost_AdmittedFalse(t *testing.T) {
	r := &PriorityBoostReconciler{
		minAdmitDuration: 30 * time.Minute,
		boostValue:       100000,
	}
	wl := &kueue.Workload{
		Status: kueue.WorkloadStatus{
			Conditions: []metav1.Condition{
				{
					Type:               string(kueue.WorkloadAdmitted),
					Status:             metav1.ConditionFalse,
					LastTransitionTime: metav1.NewTime(time.Now().Add(-5 * time.Minute)),
					Reason:             "Evicted",
				},
			},
		},
	}
	boost, requeueAfter := r.computeBoost(wl)
	if boost != 0 {
		t.Errorf("expected boost=0, got %d", boost)
	}
	if requeueAfter != 0 {
		t.Errorf("expected requeueAfter=0, got %v", requeueAfter)
	}
}

func TestComputeBoost_WithinWindow(t *testing.T) {
	const boostValue = int32(100000)
	const window = 30 * time.Minute
	admittedAt := time.Now().Add(-5 * time.Minute)

	r := &PriorityBoostReconciler{
		minAdmitDuration: window,
		boostValue:       boostValue,
	}
	wl := admittedWorkload(admittedAt)
	boost, requeueAfter := r.computeBoost(wl)
	if boost != 0 {
		t.Errorf("expected boost=0 during window, got %d", boost)
	}
	expectedRemaining := window - 5*time.Minute
	tolerance := 2 * time.Second
	if requeueAfter < expectedRemaining-tolerance || requeueAfter > expectedRemaining+tolerance {
		t.Errorf("expected requeueAfter≈%v, got %v", expectedRemaining, requeueAfter)
	}
}

func TestComputeBoost_WindowExpired_NegativeBoost(t *testing.T) {
	const boostValue = int32(100000)
	r := &PriorityBoostReconciler{
		minAdmitDuration: 30 * time.Minute,
		boostValue:       boostValue,
	}
	wl := admittedWorkload(time.Now().Add(-31 * time.Minute))
	boost, requeueAfter := r.computeBoost(wl)
	if want := -boostValue; boost != want {
		t.Errorf("expected boost=%d, got %d", want, boost)
	}
	if requeueAfter != 0 {
		t.Errorf("expected requeueAfter=0, got %v", requeueAfter)
	}
}

func TestComputeBoost_Disabled(t *testing.T) {
	r := &PriorityBoostReconciler{
		minAdmitDuration: 0,
		boostValue:       100000,
	}
	wl := admittedWorkload(time.Now().Add(-1 * time.Minute))
	boost, requeueAfter := r.computeBoost(wl)
	if boost != 0 {
		t.Errorf("expected boost=0, got %d", boost)
	}
	if requeueAfter != 0 {
		t.Errorf("expected requeueAfter=0, got %v", requeueAfter)
	}
}

func TestReconcile_WithinWindow_NoAnnotation(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	if err := kueue.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed adding kueue to scheme: %v", err)
	}

	wl := admittedWorkload(time.Now().Add(-2 * time.Minute))
	cl := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(wl).WithObjects(wl).Build()
	r := NewPriorityBoostReconciler(cl, record.NewFakeRecorder(32),
		WithMinAdmitDuration(30*time.Minute),
		WithBoostValue(100000),
	)
	req := types.NamespacedName{Name: wl.Name, Namespace: wl.Namespace}
	result, err := r.Reconcile(ctx, ctrlRequest(req))
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	if result.RequeueAfter <= 0 {
		t.Errorf("expected RequeueAfter > 0, got %v", result.RequeueAfter)
	}

	var updated kueue.Workload
	if err := cl.Get(ctx, req, &updated); err != nil {
		t.Fatalf("Get workload failed: %v", err)
	}
	if _, ok := updated.Annotations[constants.PriorityBoostAnnotationKey]; ok {
		t.Errorf("expected no annotation during window, got %q",
			updated.Annotations[constants.PriorityBoostAnnotationKey])
	}
}

func TestReconcile_AfterWindow_SetsNegativeBoost(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	if err := kueue.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed adding kueue to scheme: %v", err)
	}

	wl := admittedWorkload(time.Now().Add(-60 * time.Minute))
	wl.Annotations = map[string]string{
		constants.PriorityBoostAnnotationKey: "100000",
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(wl).WithObjects(wl).Build()
	r := NewPriorityBoostReconciler(cl, record.NewFakeRecorder(32),
		WithMinAdmitDuration(30*time.Minute),
		WithBoostValue(100000),
	)
	req := types.NamespacedName{Name: wl.Name, Namespace: wl.Namespace}
	result, err := r.Reconcile(ctx, ctrlRequest(req))
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Errorf("expected no requeue after window, got RequeueAfter=%v", result.RequeueAfter)
	}

	var updated kueue.Workload
	if err := cl.Get(ctx, req, &updated); err != nil {
		t.Fatalf("Get workload failed: %v", err)
	}
	if got, want := updated.Annotations[constants.PriorityBoostAnnotationKey], "-100000"; got != want {
		t.Errorf("expected annotation=%q, got %q", want, got)
	}
}

func TestReconcile_NotAdmitted_ClearsStaleAnnotation(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	if err := kueue.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed adding kueue to scheme: %v", err)
	}

	wl := &kueue.Workload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wl-pending",
			Namespace: "ns",
			Annotations: map[string]string{
				constants.PriorityBoostAnnotationKey: "-100000",
			},
		},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(wl).WithObjects(wl).Build()
	r := NewPriorityBoostReconciler(cl, record.NewFakeRecorder(32),
		WithMinAdmitDuration(30*time.Minute),
		WithBoostValue(100000),
	)
	req := types.NamespacedName{Name: wl.Name, Namespace: wl.Namespace}
	if _, err := r.Reconcile(ctx, ctrlRequest(req)); err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	var updated kueue.Workload
	if err := cl.Get(ctx, req, &updated); err != nil {
		t.Fatalf("Get workload failed: %v", err)
	}
	if _, ok := updated.Annotations[constants.PriorityBoostAnnotationKey]; ok {
		t.Errorf("expected annotation cleared for pending workload")
	}
}

func TestReconcile_OutOfSelector_ClearsAnnotation(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	if err := kueue.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed adding kueue to scheme: %v", err)
	}

	sel, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{"tier": "batch"},
	})
	if err != nil {
		t.Fatal(err)
	}

	wl := admittedWorkload(time.Now().Add(-60 * time.Minute))
	wl.Labels = map[string]string{"tier": "interactive"}
	wl.Annotations = map[string]string{
		constants.PriorityBoostAnnotationKey: "-100000",
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(wl).WithObjects(wl).Build()
	r := NewPriorityBoostReconciler(cl, record.NewFakeRecorder(32),
		WithMinAdmitDuration(30*time.Minute),
		WithBoostValue(100000),
		WithWorkloadSelector(sel),
	)
	req := types.NamespacedName{Name: wl.Name, Namespace: wl.Namespace}
	if _, err := r.Reconcile(ctx, ctrlRequest(req)); err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	var updated kueue.Workload
	if err := cl.Get(ctx, req, &updated); err != nil {
		t.Fatalf("Get workload failed: %v", err)
	}
	if _, ok := updated.Annotations[constants.PriorityBoostAnnotationKey]; ok {
		t.Errorf("expected annotation cleared when out of selector")
	}
}

func TestReconcile_AboveMaxPriority_ClearsAnnotation(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	if err := kueue.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed adding kueue to scheme: %v", err)
	}

	wl := admittedWorkload(time.Now().Add(-60 * time.Minute))
	prio := int32(500)
	wl.Spec.Priority = &prio
	wl.Annotations = map[string]string{
		constants.PriorityBoostAnnotationKey: "-100000",
	}
	maxP := int32(100)
	cl := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(wl).WithObjects(wl).Build()
	r := NewPriorityBoostReconciler(cl, record.NewFakeRecorder(32),
		WithMinAdmitDuration(30*time.Minute),
		WithBoostValue(100000),
		WithMaxWorkloadPriority(&maxP),
	)
	req := types.NamespacedName{Name: wl.Name, Namespace: wl.Namespace}
	if _, err := r.Reconcile(ctx, ctrlRequest(req)); err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	var updated kueue.Workload
	if err := cl.Get(ctx, req, &updated); err != nil {
		t.Fatalf("Get workload failed: %v", err)
	}
	if _, ok := updated.Annotations[constants.PriorityBoostAnnotationKey]; ok {
		t.Errorf("expected annotation cleared when above maxWorkloadPriority")
	}
}

func TestReconcile_IdempotentWithinWindow(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	if err := kueue.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed adding kueue to scheme: %v", err)
	}

	wl := admittedWorkload(time.Now().Add(-2 * time.Minute))
	patchCalled := false
	baseClient := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(wl).WithObjects(wl).Build()
	cl := interceptor.NewClient(baseClient, interceptor.Funcs{
		Patch: func(ctx context.Context, c client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
			patchCalled = true
			return c.Patch(ctx, obj, patch, opts...)
		},
	})
	r := NewPriorityBoostReconciler(cl, record.NewFakeRecorder(32),
		WithMinAdmitDuration(30*time.Minute),
		WithBoostValue(100000),
	)
	req := types.NamespacedName{Name: wl.Name, Namespace: wl.Namespace}
	if _, err := r.Reconcile(ctx, ctrlRequest(req)); err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	if patchCalled {
		t.Error("expected no patch when annotation already absent during window")
	}
}

func TestReconcile_IdempotentNegativeBoost(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	if err := kueue.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed adding kueue to scheme: %v", err)
	}

	wl := admittedWorkload(time.Now().Add(-60 * time.Minute))
	wl.Annotations = map[string]string{
		constants.PriorityBoostAnnotationKey: "-100000",
	}
	patchCalled := false
	baseClient := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(wl).WithObjects(wl).Build()
	cl := interceptor.NewClient(baseClient, interceptor.Funcs{
		Patch: func(ctx context.Context, c client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
			patchCalled = true
			return c.Patch(ctx, obj, patch, opts...)
		},
	})
	r := NewPriorityBoostReconciler(cl, record.NewFakeRecorder(32),
		WithMinAdmitDuration(30*time.Minute),
		WithBoostValue(100000),
	)
	req := types.NamespacedName{Name: wl.Name, Namespace: wl.Namespace}
	if _, err := r.Reconcile(ctx, ctrlRequest(req)); err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	if patchCalled {
		t.Error("expected no patch when annotation already correct")
	}
}

func TestReconcile_ConflictThenSuccess(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	_ = kueue.AddToScheme(scheme)

	// Post-window: controller must patch to set negative boost; first patch conflicts.
	wl := admittedWorkload(time.Now().Add(-60 * time.Minute))
	baseClient := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(wl).WithObjects(wl).Build()
	conflictInjected := false
	cl := interceptor.NewClient(baseClient, interceptor.Funcs{
		Patch: func(ctx context.Context, c client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
			if !conflictInjected {
				conflictInjected = true
				return apierrors.NewConflict(schema.GroupResource{Group: "kueue.x-k8s.io", Resource: "workloads"}, obj.GetName(), nil)
			}
			return c.Patch(ctx, obj, patch, opts...)
		},
	})

	r := NewPriorityBoostReconciler(cl, record.NewFakeRecorder(32),
		WithMinAdmitDuration(30*time.Minute),
		WithBoostValue(100000),
	)
	req := types.NamespacedName{Name: wl.Name, Namespace: wl.Namespace}
	_, err := r.Reconcile(ctx, ctrlRequest(req))
	if err == nil {
		t.Fatal("expected first reconcile to fail with conflict")
	}
	_, err = r.Reconcile(ctx, ctrlRequest(req))
	if err != nil {
		t.Fatalf("expected second reconcile to succeed, got: %v", err)
	}

	var updated kueue.Workload
	if err := baseClient.Get(ctx, req, &updated); err != nil {
		t.Fatalf("Get workload failed: %v", err)
	}
	if got, want := updated.Annotations[constants.PriorityBoostAnnotationKey], "-100000"; got != want {
		t.Errorf("expected annotation=%q, got %q", want, got)
	}
}

func ctrlRequest(name types.NamespacedName) ctrl.Request {
	return ctrl.Request{NamespacedName: name}
}
