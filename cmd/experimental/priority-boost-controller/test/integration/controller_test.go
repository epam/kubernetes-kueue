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

package integration

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kueue "sigs.k8s.io/kueue/apis/kueue/v1beta2"
	"sigs.k8s.io/kueue/cmd/experimental/priority-boost-controller/pkg/constants"
	controllerpkg "sigs.k8s.io/kueue/cmd/experimental/priority-boost-controller/pkg/controller"
)

// TestReconcile_NoAnnotationWithinWindow verifies that when a workload has been admitted
// for less than minAdmitDuration, no priority-boost annotation is set and RequeueAfter is positive.
func TestReconcile_NoAnnotationWithinWindow(t *testing.T) {
	ctx := t.Context()
	scheme := runtime.NewScheme()
	if err := kueue.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed adding kueue to scheme: %v", err)
	}

	admittedAt := time.Now().Add(-5 * time.Minute)
	wl := &kueue.Workload{
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

	cl := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(wl).WithObjects(wl).Build()
	reconciler := controllerpkg.NewPriorityBoostReconciler(
		cl,
		record.NewFakeRecorder(32),
		controllerpkg.WithMinAdmitDuration(30*time.Minute),
		controllerpkg.WithBoostValue(100000),
	)

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{Name: wl.Name, Namespace: wl.Namespace},
	}
	result, err := reconciler.Reconcile(ctx, req)
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	if result.RequeueAfter <= 0 {
		t.Errorf("expected RequeueAfter > 0, got %v", result.RequeueAfter)
	}

	var updated kueue.Workload
	if err := cl.Get(ctx, types.NamespacedName{Name: wl.Name, Namespace: wl.Namespace}, &updated); err != nil {
		t.Fatalf("Get workload failed: %v", err)
	}
	if _, ok := updated.Annotations[constants.PriorityBoostAnnotationKey]; ok {
		t.Errorf("expected no annotation during window, got %q",
			updated.Annotations[constants.PriorityBoostAnnotationKey])
	}
}

// TestReconcile_NegativeBoostAfterWindow verifies that after minAdmitDuration the annotation is -boostValue.
func TestReconcile_NegativeBoostAfterWindow(t *testing.T) {
	ctx := t.Context()
	scheme := runtime.NewScheme()
	if err := kueue.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed adding kueue to scheme: %v", err)
	}

	admittedAt := time.Now().Add(-60 * time.Minute)
	wl := &kueue.Workload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wl-1",
			Namespace: "ns",
			Annotations: map[string]string{
				constants.PriorityBoostAnnotationKey: "100000",
			},
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

	cl := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(wl).WithObjects(wl).Build()
	reconciler := controllerpkg.NewPriorityBoostReconciler(
		cl,
		record.NewFakeRecorder(32),
		controllerpkg.WithMinAdmitDuration(30*time.Minute),
		controllerpkg.WithBoostValue(100000),
	)

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{Name: wl.Name, Namespace: wl.Namespace},
	}
	result, err := reconciler.Reconcile(ctx, req)
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Errorf("expected no requeue after window, got RequeueAfter=%v", result.RequeueAfter)
	}

	var updated kueue.Workload
	if err := cl.Get(ctx, types.NamespacedName{Name: wl.Name, Namespace: wl.Namespace}, &updated); err != nil {
		t.Fatalf("Get workload failed: %v", err)
	}
	if got, want := updated.Annotations[constants.PriorityBoostAnnotationKey], "-100000"; got != want {
		t.Errorf("expected annotation=%q, got %q", want, got)
	}
}

// TestReconcile_NoBoostWhenNotAdmitted verifies that a pending workload does not receive a priority-boost annotation.
func TestReconcile_NoBoostWhenNotAdmitted(t *testing.T) {
	ctx := t.Context()
	scheme := runtime.NewScheme()
	if err := kueue.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed adding kueue to scheme: %v", err)
	}

	wl := &kueue.Workload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wl-pending",
			Namespace: "ns",
		},
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(wl).WithObjects(wl).Build()
	reconciler := controllerpkg.NewPriorityBoostReconciler(
		cl,
		record.NewFakeRecorder(32),
		controllerpkg.WithMinAdmitDuration(30*time.Minute),
		controllerpkg.WithBoostValue(100000),
	)

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{Name: wl.Name, Namespace: wl.Namespace},
	}
	result, err := reconciler.Reconcile(ctx, req)
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Errorf("expected no requeue for pending workload, got RequeueAfter=%v", result.RequeueAfter)
	}

	var updated kueue.Workload
	if err := cl.Get(ctx, types.NamespacedName{Name: wl.Name, Namespace: wl.Namespace}, &updated); err != nil {
		t.Fatalf("Get workload failed: %v", err)
	}
	if _, ok := updated.Annotations[constants.PriorityBoostAnnotationKey]; ok {
		t.Errorf("expected no annotation on pending workload, got: %q",
			updated.Annotations[constants.PriorityBoostAnnotationKey])
	}
}
