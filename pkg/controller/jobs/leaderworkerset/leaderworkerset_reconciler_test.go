/*
Copyright 2024 The Kubernetes Authors.

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
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	leaderworkersetv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"

	podcontroller "sigs.k8s.io/kueue/pkg/controller/jobs/pod"
	utiltesting "sigs.k8s.io/kueue/pkg/util/testing"
	"sigs.k8s.io/kueue/pkg/util/testingjobs/leaderworkerset"
	testingjobspod "sigs.k8s.io/kueue/pkg/util/testingjobs/pod"
)

var (
	baseCmpOpts = []cmp.Option{
		cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(metav1.ObjectMeta{}, "ResourceVersion"),
	}
)

func TestReconciler(t *testing.T) {
	now := time.Now()

	cases := map[string]struct {
		lws      *leaderworkersetv1.LeaderWorkerSet
		pods     []corev1.Pod
		wantLWS  *leaderworkersetv1.LeaderWorkerSet
		wantPods []corev1.Pod
		wantErr  error
	}{
		"should add finalizer": {
			lws:     leaderworkerset.MakeLeaderWorkerSet(testLWSName, testNSName).Obj(),
			wantLWS: leaderworkerset.MakeLeaderWorkerSet(testLWSName, testNSName).KueueFinalizer().Obj(),
		},
		"should delete finalizer": {
			lws: leaderworkerset.MakeLeaderWorkerSet(testLWSName, testNSName).
				DeletionTimestamp(now).
				KueueFinalizer().
				Obj(),
		},
		"should finalize pods on deleting leaderworkerset": {
			lws: leaderworkerset.MakeLeaderWorkerSet(testLWSName, testNSName).
				DeletionTimestamp(now).
				KueueFinalizer().
				Obj(),
			pods: []corev1.Pod{
				*testingjobspod.MakePod("pod1", testNSName).
					Label(leaderworkersetv1.SetNameLabelKey, testLWSName).
					KueueFinalizer().
					Obj(),
			},
			wantPods: []corev1.Pod{
				*testingjobspod.MakePod("pod1", testNSName).
					Label(leaderworkersetv1.SetNameLabelKey, testLWSName).
					Obj(),
			},
		},
		"should finalize succeeded and failed pods": {
			lws: leaderworkerset.MakeLeaderWorkerSet(testLWSName, testNSName).KueueFinalizer().Obj(),
			pods: []corev1.Pod{
				*testingjobspod.MakePod("pod1", testNSName).
					Label(leaderworkersetv1.SetNameLabelKey, testLWSName).
					StatusPhase(corev1.PodSucceeded).
					KueueFinalizer().
					Obj(),
				*testingjobspod.MakePod("pod2", testNSName).
					Label(leaderworkersetv1.SetNameLabelKey, testLWSName).
					StatusPhase(corev1.PodFailed).
					KueueFinalizer().
					Obj(),
				*testingjobspod.MakePod("pod3", testNSName).
					Label(leaderworkersetv1.SetNameLabelKey, testLWSName).
					KueueFinalizer().
					Obj(),
			},
			wantLWS: leaderworkerset.MakeLeaderWorkerSet(testLWSName, testNSName).KueueFinalizer().Obj(),
			wantPods: []corev1.Pod{
				*testingjobspod.MakePod("pod1", testNSName).
					Label(leaderworkersetv1.SetNameLabelKey, testLWSName).
					StatusPhase(corev1.PodSucceeded).
					Obj(),
				*testingjobspod.MakePod("pod2", testNSName).
					Label(leaderworkersetv1.SetNameLabelKey, testLWSName).
					StatusPhase(corev1.PodFailed).
					Obj(),
				*testingjobspod.MakePod("pod3", testNSName).
					Label(leaderworkersetv1.SetNameLabelKey, testLWSName).
					KueueFinalizer().
					Obj(),
			},
		},
		"shouldn't set default values without group index label": {
			lws: leaderworkerset.MakeLeaderWorkerSet(testLWSName, testNSName).KueueFinalizer().Obj(),
			pods: []corev1.Pod{
				*testingjobspod.MakePod("pod1", testNSName).
					Label(leaderworkersetv1.SetNameLabelKey, testLWSName).
					Obj(),
			},
			wantLWS: leaderworkerset.MakeLeaderWorkerSet(testLWSName, testNSName).KueueFinalizer().Obj(),
			wantPods: []corev1.Pod{
				*testingjobspod.MakePod("pod1", testNSName).
					Label(leaderworkersetv1.SetNameLabelKey, testLWSName).
					Obj(),
			},
		},
		"should set default values": {
			lws: leaderworkerset.MakeLeaderWorkerSet(testLWSName, testNSName).
				Queue("queue").
				KueueFinalizer().
				Obj(),
			pods: []corev1.Pod{
				*testingjobspod.MakePod("pod1", testNSName).
					Label(leaderworkersetv1.SetNameLabelKey, testLWSName).
					Label(leaderworkersetv1.GroupIndexLabelKey, "0").
					Annotation(podcontroller.SuspendedByParentAnnotation, FrameworkName).
					Annotation(podcontroller.GroupServingAnnotation, "true").
					Obj(),
			},
			wantLWS: leaderworkerset.MakeLeaderWorkerSet(testLWSName, testNSName).
				Queue("queue").
				KueueFinalizer().
				Obj(),
			wantPods: []corev1.Pod{
				*testingjobspod.MakePod("pod1", testNSName).
					Label(leaderworkersetv1.SetNameLabelKey, testLWSName).
					Label(leaderworkersetv1.GroupIndexLabelKey, "0").
					Queue("queue").
					Group("leaderworkerset-test-lws-7aa6c7b8-0-769be").
					GroupTotalCount("1").
					Annotation(podcontroller.SuspendedByParentAnnotation, FrameworkName).
					Annotation(podcontroller.GroupServingAnnotation, "true").
					Annotation(podcontroller.RoleHashAnnotation, "7aa6c7b8").
					Obj(),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, _ := utiltesting.ContextWithLog(t)
			clientBuilder := utiltesting.NewClientBuilder(leaderworkersetv1.AddToScheme)

			initObjs := make([]client.Object, 0, len(tc.pods)+1)
			initObjs = append(initObjs, tc.lws)
			for _, pod := range tc.pods {
				initObjs = append(initObjs, &pod)
			}

			kClient := clientBuilder.WithObjects(initObjs...).Build()

			reconciler := NewReconciler(kClient, nil)

			lwsKey := client.ObjectKeyFromObject(tc.lws)
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: lwsKey})
			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Reconcile returned error (-want,+got):\n%s", diff)
			}

			gotLWS := &leaderworkersetv1.LeaderWorkerSet{}
			err = kClient.Get(ctx, lwsKey, gotLWS)
			if err != nil {
				if !errors.IsNotFound(err) {
					t.Fatalf("Could not get LeaderWorkerSet after reconcile: %v", err)
				}
				gotLWS = nil
			}

			if diff := cmp.Diff(tc.wantLWS, gotLWS, baseCmpOpts...); diff != "" {
				t.Errorf("LeaderWorkerSet after reconcile (-want,+got):\n%s", diff)
			}

			gotPodList := &corev1.PodList{}
			if err := kClient.List(ctx, gotPodList); err != nil {
				t.Fatalf("Could not get Pods after reconcile: %v", err)
			}

			if diff := cmp.Diff(tc.wantPods, gotPodList.Items, baseCmpOpts...); diff != "" {
				t.Errorf("Pods after reconcile (-want,+got):\n%s", diff)
			}
		})
	}
}
