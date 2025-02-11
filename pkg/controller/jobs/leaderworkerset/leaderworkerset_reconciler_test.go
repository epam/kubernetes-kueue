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
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	leaderworkersetv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"

	kueue "sigs.k8s.io/kueue/apis/kueue/v1beta1"
	"sigs.k8s.io/kueue/pkg/controller/jobframework"
	podcontroller "sigs.k8s.io/kueue/pkg/controller/jobs/pod"
	utiltesting "sigs.k8s.io/kueue/pkg/util/testing"
	"sigs.k8s.io/kueue/pkg/util/testingjobs/leaderworkerset"
	testingjobspod "sigs.k8s.io/kueue/pkg/util/testingjobs/pod"
	"sigs.k8s.io/kueue/pkg/util/testingjobs/statefulset"
)

var (
	baseCmpOpts = []cmp.Option{
		cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(metav1.ObjectMeta{}, "ResourceVersion"),
	}
)

var (
	testNS  = "test-ns"
	testLWS = "test-lws"
	testUID = "test-uid"
)

func TestReconciler(t *testing.T) {
	cases := map[string]struct {
		lwsKey              client.ObjectKey
		labelKeysToCopy     []string
		leaderWorkerSet     *leaderworkersetv1.LeaderWorkerSet
		statefulSets        []appsv1.StatefulSet
		pods                []corev1.Pod
		workloads           []kueue.Workload
		wantLeaderWorkerSet *leaderworkersetv1.LeaderWorkerSet
		wantPods            []corev1.Pod
		wantWorkloads       []kueue.Workload
		wantEvents          []utiltesting.EventRecord
		wantErr             error
	}{
		"leaderworkerset with finished pods": {
			lwsKey: client.ObjectKey{Name: testLWS, Namespace: testNS},
			statefulSets: []appsv1.StatefulSet{
				*statefulset.MakeStatefulSet("sts1", testNS).Obj(),
			},
			pods: []corev1.Pod{
				*testingjobspod.MakePod("pod1-1", testNS).
					OwnerReference("sts1", appsv1.SchemeGroupVersion.WithKind("StatefulSet")).
					Label(leaderworkersetv1.SetNameLabelKey, testLWS).
					KueueFinalizer().
					StatusPhase(corev1.PodSucceeded).
					Obj(),
				*testingjobspod.MakePod("pod1-2", testNS).
					OwnerReference("sts1", appsv1.SchemeGroupVersion.WithKind("StatefulSet")).
					Label(leaderworkersetv1.SetNameLabelKey, testLWS).
					KueueFinalizer().
					StatusPhase(corev1.PodFailed).
					Obj(),
				*testingjobspod.MakePod("pod1-3", testNS).
					OwnerReference("sts1", appsv1.SchemeGroupVersion.WithKind("StatefulSet")).
					Label(leaderworkersetv1.SetNameLabelKey, testLWS).
					KueueFinalizer().
					Obj(),
				*testingjobspod.MakePod("pod2-1", testNS).
					OwnerReference("sts2", appsv1.SchemeGroupVersion.WithKind("StatefulSet")).
					Label(leaderworkersetv1.SetNameLabelKey, testLWS).
					KueueFinalizer().
					StatusPhase(corev1.PodSucceeded).
					Obj(),
				*testingjobspod.MakePod("pod2-2", testNS).
					OwnerReference("sts2", appsv1.SchemeGroupVersion.WithKind("StatefulSet")).
					Label(leaderworkersetv1.SetNameLabelKey, testLWS).
					KueueFinalizer().
					StatusPhase(corev1.PodFailed).
					Obj(),
				*testingjobspod.MakePod("pod2-3", testNS).
					OwnerReference("sts2", appsv1.SchemeGroupVersion.WithKind("StatefulSet")).
					Label(leaderworkersetv1.SetNameLabelKey, testLWS).
					KueueFinalizer().
					Obj(),
				*testingjobspod.MakePod("pod3-no-owner", testNS).
					Label(leaderworkersetv1.SetNameLabelKey, testLWS).
					KueueFinalizer().
					Obj(),
			},
			wantPods: []corev1.Pod{
				*testingjobspod.MakePod("pod1-1", testNS).
					OwnerReference("sts1", appsv1.SchemeGroupVersion.WithKind("StatefulSet")).
					Label(leaderworkersetv1.SetNameLabelKey, testLWS).
					StatusPhase(corev1.PodSucceeded).
					Obj(),
				*testingjobspod.MakePod("pod1-2", testNS).
					OwnerReference("sts1", appsv1.SchemeGroupVersion.WithKind("StatefulSet")).
					Label(leaderworkersetv1.SetNameLabelKey, testLWS).
					StatusPhase(corev1.PodFailed).
					Obj(),
				*testingjobspod.MakePod("pod1-3", testNS).
					OwnerReference("sts1", appsv1.SchemeGroupVersion.WithKind("StatefulSet")).
					Label(leaderworkersetv1.SetNameLabelKey, testLWS).
					KueueFinalizer().
					Obj(),
				*testingjobspod.MakePod("pod2-1", testNS).
					OwnerReference("sts2", appsv1.SchemeGroupVersion.WithKind("StatefulSet")).
					Label(leaderworkersetv1.SetNameLabelKey, testLWS).
					StatusPhase(corev1.PodSucceeded).
					Obj(),
				*testingjobspod.MakePod("pod2-2", testNS).
					OwnerReference("sts2", appsv1.SchemeGroupVersion.WithKind("StatefulSet")).
					Label(leaderworkersetv1.SetNameLabelKey, testLWS).
					StatusPhase(corev1.PodFailed).
					Obj(),
				*testingjobspod.MakePod("pod2-3", testNS).
					OwnerReference("sts2", appsv1.SchemeGroupVersion.WithKind("StatefulSet")).
					Label(leaderworkersetv1.SetNameLabelKey, testLWS).
					Obj(),
				*testingjobspod.MakePod("pod3-no-owner", testNS).
					Label(leaderworkersetv1.SetNameLabelKey, testLWS).
					Obj(),
			},
		},
		"should create prebuild workload": {
			lwsKey:              client.ObjectKey{Name: testLWS, Namespace: testNS},
			leaderWorkerSet:     leaderworkerset.MakeLeaderWorkerSet(testLWS, testNS).UID(testUID).Obj(),
			wantLeaderWorkerSet: leaderworkerset.MakeLeaderWorkerSet(testLWS, testNS).UID(testUID).Obj(),
			wantWorkloads: []kueue.Workload{
				*utiltesting.MakeWorkload(GetWorkloadName(types.UID(testUID), testLWS, "0"), testNS).
					Annotation(podcontroller.IsGroupWorkloadAnnotationKey, podcontroller.IsGroupWorkloadAnnotationValue).
					Finalizers(kueue.ResourceInUseFinalizerName).
					PodSets(
						kueue.PodSet{
							Name: kueue.DefaultPodSetName,
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{Name: "c", Image: "pause"},
									},
								},
							},
							Count:           1,
							TopologyRequest: nil,
						}).
					Obj(),
			},
			wantEvents: []utiltesting.EventRecord{
				{
					Key:       types.NamespacedName{Name: testLWS, Namespace: testNS},
					EventType: corev1.EventTypeNormal,
					Reason:    jobframework.ReasonCreatedWorkload,
					Message: fmt.Sprintf(
						"Created Workload: %s/%s",
						testNS,
						GetWorkloadName(types.UID(testUID), testLWS, "0"),
					),
				},
			},
		},
		"should create prebuild workload with leader template": {
			lwsKey: client.ObjectKey{Name: testLWS, Namespace: testNS},
			leaderWorkerSet: leaderworkerset.MakeLeaderWorkerSet(testLWS, testNS).
				UID(testUID).
				Size(3).
				LeaderTemplate(corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "c", Image: "pause"},
						},
					},
				}).
				Obj(),
			wantLeaderWorkerSet: leaderworkerset.MakeLeaderWorkerSet(testLWS, testNS).
				UID(testUID).
				Size(3).
				LeaderTemplate(corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "c", Image: "pause"},
						},
					},
				}).
				Obj(),
			wantWorkloads: []kueue.Workload{
				*utiltesting.MakeWorkload(GetWorkloadName(types.UID(testUID), testLWS, "0"), testNS).
					Annotation(podcontroller.IsGroupWorkloadAnnotationKey, podcontroller.IsGroupWorkloadAnnotationValue).
					Finalizers(kueue.ResourceInUseFinalizerName).
					PodSets(
						kueue.PodSet{
							Name: leaderPodSetName,
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{Name: "c", Image: "pause"},
									},
								},
							},
							Count:           1,
							TopologyRequest: nil,
						},
						kueue.PodSet{
							Name: workerPodSetName,
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{Name: "c", Image: "pause"},
									},
								},
							},
							Count:           2,
							TopologyRequest: nil,
						},
					).
					Obj(),
			},
			wantEvents: []utiltesting.EventRecord{
				{
					Key:       types.NamespacedName{Name: testLWS, Namespace: testNS},
					EventType: corev1.EventTypeNormal,
					Reason:    jobframework.ReasonCreatedWorkload,
					Message: fmt.Sprintf(
						"Created Workload: %s/%s",
						testNS,
						GetWorkloadName(types.UID(testUID), testLWS, "0"),
					),
				},
			},
		},
		"should create prebuild workloads with leader template": {
			lwsKey: client.ObjectKey{Name: testLWS, Namespace: testNS},
			leaderWorkerSet: leaderworkerset.MakeLeaderWorkerSet(testLWS, testNS).
				UID(testUID).
				Replicas(2).
				Size(3).
				LeaderTemplate(corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "c", Image: "pause"},
						},
					},
				}).
				Obj(),
			wantLeaderWorkerSet: leaderworkerset.MakeLeaderWorkerSet(testLWS, testNS).
				UID(testUID).
				Replicas(2).
				Size(3).
				LeaderTemplate(corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "c", Image: "pause"},
						},
					},
				}).
				Obj(),
			wantWorkloads: []kueue.Workload{
				*utiltesting.MakeWorkload(GetWorkloadName(types.UID(testUID), testLWS, "0"), testNS).
					Annotation(podcontroller.IsGroupWorkloadAnnotationKey, podcontroller.IsGroupWorkloadAnnotationValue).
					Finalizers(kueue.ResourceInUseFinalizerName).
					PodSets(
						kueue.PodSet{
							Name: leaderPodSetName,
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{Name: "c", Image: "pause"},
									},
								},
							},
							Count:           1,
							TopologyRequest: nil,
						},
						kueue.PodSet{
							Name: workerPodSetName,
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{Name: "c", Image: "pause"},
									},
								},
							},
							Count:           2,
							TopologyRequest: nil,
						},
					).
					Obj(),
				*utiltesting.MakeWorkload(GetWorkloadName(types.UID(testUID), testLWS, "1"), testNS).
					Annotation(podcontroller.IsGroupWorkloadAnnotationKey, podcontroller.IsGroupWorkloadAnnotationValue).
					Finalizers(kueue.ResourceInUseFinalizerName).
					PodSets(
						kueue.PodSet{
							Name: leaderPodSetName,
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{Name: "c", Image: "pause"},
									},
								},
							},
							Count:           1,
							TopologyRequest: nil,
						},
						kueue.PodSet{
							Name: workerPodSetName,
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{Name: "c", Image: "pause"},
									},
								},
							},
							Count:           2,
							TopologyRequest: nil,
						},
					).
					Obj(),
			},
			wantEvents: []utiltesting.EventRecord{
				{
					Key:       types.NamespacedName{Name: testLWS, Namespace: testNS},
					EventType: corev1.EventTypeNormal,
					Reason:    jobframework.ReasonCreatedWorkload,
					Message: fmt.Sprintf(
						"Created Workload: %s/%s",
						testNS,
						GetWorkloadName(types.UID(testUID), testLWS, "0"),
					),
				},
				{
					Key:       types.NamespacedName{Name: testLWS, Namespace: testNS},
					EventType: corev1.EventTypeNormal,
					Reason:    jobframework.ReasonCreatedWorkload,
					Message: fmt.Sprintf(
						"Created Workload: %s/%s",
						testNS,
						GetWorkloadName(types.UID(testUID), testLWS, "1"),
					),
				},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, _ := utiltesting.ContextWithLog(t)
			clientBuilder := utiltesting.NewClientBuilder(leaderworkersetv1.AddToScheme)

			objs := make([]client.Object, 0, len(tc.statefulSets)+len(tc.pods)+len(tc.workloads)+1)
			if tc.leaderWorkerSet != nil {
				objs = append(objs, tc.leaderWorkerSet)
			}
			for _, sts := range tc.statefulSets {
				objs = append(objs, &sts)
			}
			for _, pod := range tc.pods {
				objs = append(objs, &pod)
			}
			for _, wl := range tc.workloads {
				objs = append(objs, &wl)
			}

			kClient := clientBuilder.WithObjects(objs...).Build()
			recorder := &utiltesting.EventRecorder{}

			reconciler := NewReconciler(kClient, recorder, jobframework.WithLabelKeysToCopy(tc.labelKeysToCopy))

			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: tc.lwsKey})
			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Reconcile returned error (-want,+got):\n%s", diff)
			}

			gotLeaderWorkerSet := &leaderworkersetv1.LeaderWorkerSet{}
			if err := kClient.Get(ctx, tc.lwsKey, gotLeaderWorkerSet); err != nil {
				if !errors.IsNotFound(err) {
					t.Fatalf("Could not get LeaderWorkerSet after reconcile: %v", err)
				}
				gotLeaderWorkerSet = nil
			}

			if diff := cmp.Diff(tc.wantLeaderWorkerSet, gotLeaderWorkerSet, baseCmpOpts...); diff != "" {
				t.Errorf("LeaderWorkerSet after reconcile (-want,+got):\n%s", diff)
			}

			gotWorkloads := kueue.WorkloadList{}
			err = kClient.List(ctx, &gotWorkloads, client.InNamespace(tc.lwsKey.Namespace))
			if err != nil {
				t.Fatalf("Could not get Workloads after reconcile: %v", err)
			}

			if diff := cmp.Diff(tc.wantWorkloads, gotWorkloads.Items, baseCmpOpts...); diff != "" {
				t.Errorf("Workloads after reconcile (-want,+got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.wantEvents, recorder.RecordedEvents, cmpopts.SortSlices(utiltesting.SortEvents)); diff != "" {
				t.Errorf("Unexpected events (-want/+got):\n%s", diff)
			}

			gotPodList := &corev1.PodList{}
			if err := kClient.List(ctx, gotPodList); err != nil {
				t.Fatalf("Could not get PodList after reconcile: %v", err)
			}

			if diff := cmp.Diff(tc.wantPods, gotPodList.Items, baseCmpOpts...); diff != "" {
				t.Errorf("Pods after reconcile (-want,+got):\n%s", diff)
			}
		})
	}
}
