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

// TestComputeBoost_NotAdmitted: workload has no Admitted condition — no boost.
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

// TestComputeBoost_AdmittedFalse: workload Admitted=False — no boost.
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

// TestComputeBoost_WithinWindow: admitted 5 min ago, window=30m → boost=boostValue, requeue ~25m.
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
	if boost != boostValue {
		t.Errorf("expected boost=%d, got %d", boostValue, boost)
	}
	// requeueAfter should be approximately 25 minutes (within a small tolerance)
	expectedRemaining := window - 5*time.Minute
	tolerance := 2 * time.Second
	if requeueAfter < expectedRemaining-tolerance || requeueAfter > expectedRemaining+tolerance {
		t.Errorf("expected requeueAfter≈%v, got %v", expectedRemaining, requeueAfter)
	}
}

// TestComputeBoost_WindowExpired: admitted 31 min ago, window=30m → no boost.
func TestComputeBoost_WindowExpired(t *testing.T) {
	r := &PriorityBoostReconciler{
		minAdmitDuration: 30 * time.Minute,
		boostValue:       100000,
	}
	wl := admittedWorkload(time.Now().Add(-31 * time.Minute))
	boost, requeueAfter := r.computeBoost(wl)
	if boost != 0 {
		t.Errorf("expected boost=0, got %d", boost)
	}
	if requeueAfter != 0 {
		t.Errorf("expected requeueAfter=0, got %v", requeueAfter)
	}
}

// TestComputeBoost_Disabled: minAdmitDuration=0 (disabled) → no boost regardless of admit time.
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

// TestReconcile_SetsBoostAnnotation: reconcile on a fresh admitted workload sets the annotation.
func TestReconcile_SetsBoostAnnotation(t *testing.T) {
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
	if got := updated.Annotations[constants.PriorityBoostAnnotationKey]; got != "100000" {
		t.Errorf("expected annotation=%q, got %q", "100000", got)
	}
}

// TestReconcile_ClearsBoostAnnotation: reconcile on an expired-window workload removes annotation.
func TestReconcile_ClearsBoostAnnotation(t *testing.T) {
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
		t.Errorf("expected no requeue, got RequeueAfter=%v", result.RequeueAfter)
	}

	var updated kueue.Workload
	if err := cl.Get(ctx, req, &updated); err != nil {
		t.Fatalf("Get workload failed: %v", err)
	}
	if _, ok := updated.Annotations[constants.PriorityBoostAnnotationKey]; ok {
		t.Errorf("expected annotation to be absent, but it is present: %q",
			updated.Annotations[constants.PriorityBoostAnnotationKey])
	}
}

// TestReconcile_Idempotent: reconcile when annotation already has the correct value makes no patch.
func TestReconcile_Idempotent(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	if err := kueue.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed adding kueue to scheme: %v", err)
	}

	wl := admittedWorkload(time.Now().Add(-2 * time.Minute))
	wl.Annotations = map[string]string{
		constants.PriorityBoostAnnotationKey: "100000",
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
		t.Error("expected no patch when annotation is already correct")
	}
}

// TestReconcile_ConflictThenSuccess: first patch returns conflict, second succeeds.
func TestReconcile_ConflictThenSuccess(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	_ = kueue.AddToScheme(scheme)

	wl := admittedWorkload(time.Now().Add(-2 * time.Minute))
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
	if got := updated.Annotations[constants.PriorityBoostAnnotationKey]; got != "100000" {
		t.Errorf("expected annotation=%q, got %q", "100000", got)
	}
}

func ctrlRequest(name types.NamespacedName) ctrl.Request {
	return ctrl.Request{NamespacedName: name}
}
