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

package core

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kueue "sigs.k8s.io/kueue/apis/kueue/v1beta2"
	"sigs.k8s.io/kueue/pkg/workload"
)

func TestIsPreemptionEvictionToPending(t *testing.T) {
	evictTime := time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)
	now := evictTime.Add(30 * time.Second)

	preemptedEvicted := metav1.Condition{
		Type:               kueue.WorkloadEvicted,
		Status:             metav1.ConditionTrue,
		Reason:             kueue.WorkloadEvictedByPreemption,
		LastTransitionTime: metav1.NewTime(evictTime),
	}
	otherEvicted := metav1.Condition{
		Type:               kueue.WorkloadEvicted,
		Status:             metav1.ConditionTrue,
		Reason:             kueue.WorkloadEvictedByPodsReadyTimeout,
		LastTransitionTime: metav1.NewTime(evictTime),
	}
	admittedTrue := metav1.Condition{
		Type:               kueue.WorkloadAdmitted,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.NewTime(evictTime),
	}
	quotaReservedTrue := metav1.Condition{
		Type:               kueue.WorkloadQuotaReserved,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.NewTime(evictTime),
	}

	cases := []struct {
		name        string
		oldWl       *kueue.Workload
		newWl       *kueue.Workload
		wantOK      bool
		wantCQ      kueue.ClusterQueueReference
		wantLatency time.Duration
	}{
		{
			name: "preemption eviction: admitted to pending, cq from old admission",
			oldWl: &kueue.Workload{
				Status: kueue.WorkloadStatus{
					Admission:  &kueue.Admission{ClusterQueue: "cq-a"},
					Conditions: []metav1.Condition{admittedTrue},
				},
			},
			newWl: &kueue.Workload{
				Status: kueue.WorkloadStatus{
					Conditions: []metav1.Condition{preemptedEvicted},
				},
			},
			wantOK:      true,
			wantCQ:      "cq-a",
			wantLatency: 30 * time.Second,
		},
		{
			name: "preemption eviction: quota reserved to pending",
			oldWl: &kueue.Workload{
				Status: kueue.WorkloadStatus{
					Admission:  &kueue.Admission{ClusterQueue: "cq-b"},
					Conditions: []metav1.Condition{quotaReservedTrue},
				},
			},
			newWl: &kueue.Workload{
				Status: kueue.WorkloadStatus{
					Conditions: []metav1.Condition{preemptedEvicted},
				},
			},
			wantOK:      true,
			wantCQ:      "cq-b",
			wantLatency: 30 * time.Second,
		},
		{
			name: "skip when old admission missing",
			oldWl: &kueue.Workload{
				Status: kueue.WorkloadStatus{},
			},
			newWl: &kueue.Workload{
				Status: kueue.WorkloadStatus{
					Conditions: []metav1.Condition{preemptedEvicted},
				},
			},
			wantOK: false,
		},
		{
			name: "skip when cluster queue empty string",
			oldWl: &kueue.Workload{
				Status: kueue.WorkloadStatus{
					Admission:  &kueue.Admission{ClusterQueue: ""},
					Conditions: []metav1.Condition{admittedTrue},
				},
			},
			newWl: &kueue.Workload{
				Status: kueue.WorkloadStatus{
					Conditions: []metav1.Condition{preemptedEvicted},
				},
			},
			wantOK: false,
		},
		{
			name: "skip wrong eviction reason",
			oldWl: &kueue.Workload{
				Status: kueue.WorkloadStatus{
					Admission:  &kueue.Admission{ClusterQueue: "cq-a"},
					Conditions: []metav1.Condition{admittedTrue},
				},
			},
			newWl: &kueue.Workload{
				Status: kueue.WorkloadStatus{
					Conditions: []metav1.Condition{otherEvicted},
				},
			},
			wantOK: false,
		},
		{
			name: "skip when new status not pending",
			oldWl: &kueue.Workload{
				Status: kueue.WorkloadStatus{
					Admission:  &kueue.Admission{ClusterQueue: "cq-a"},
					Conditions: []metav1.Condition{admittedTrue},
				},
			},
			newWl: &kueue.Workload{
				Status: kueue.WorkloadStatus{
					Admission:  &kueue.Admission{ClusterQueue: "cq-a"},
					Conditions: []metav1.Condition{admittedTrue, preemptedEvicted},
				},
			},
			wantOK: false,
		},
		{
			name: "skip when prev status pending",
			oldWl: &kueue.Workload{
				Status: kueue.WorkloadStatus{
					Conditions: []metav1.Condition{preemptedEvicted},
				},
			},
			newWl: &kueue.Workload{
				Status: kueue.WorkloadStatus{
					Conditions: []metav1.Condition{preemptedEvicted},
				},
			},
			wantOK: false,
		},
		{
			name: "skip missing eviction condition",
			oldWl: &kueue.Workload{
				Status: kueue.WorkloadStatus{
					Admission:  &kueue.Admission{ClusterQueue: "cq-a"},
					Conditions: []metav1.Condition{admittedTrue},
				},
			},
			newWl:  &kueue.Workload{Status: kueue.WorkloadStatus{}},
			wantOK: false,
		},
		{
			name: "skip eviction condition not true",
			oldWl: &kueue.Workload{
				Status: kueue.WorkloadStatus{
					Admission:  &kueue.Admission{ClusterQueue: "cq-a"},
					Conditions: []metav1.Condition{admittedTrue},
				},
			},
			newWl: &kueue.Workload{
				Status: kueue.WorkloadStatus{
					Conditions: []metav1.Condition{
						{
							Type:               kueue.WorkloadEvicted,
							Status:             metav1.ConditionFalse,
							Reason:             kueue.WorkloadEvictedByPreemption,
							LastTransitionTime: metav1.NewTime(evictTime),
						},
					},
				},
			},
			wantOK: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok := isPreemptionEvictionToPending(tc.oldWl, tc.newWl)
			if ok != tc.wantOK {
				t.Fatalf("ok: got %v want %v", ok, tc.wantOK)
			}
			if !tc.wantOK {
				return
			}
			cq, cqOK := workload.WorkloadClusterQueue(tc.oldWl)
			if !cqOK {
				t.Fatal("expected cluster queue on old workload")
			}
			if cq != tc.wantCQ {
				t.Errorf("cluster_queue: got %q want %q", cq, tc.wantCQ)
			}
			c := apimeta.FindStatusCondition(tc.newWl.Status.Conditions, kueue.WorkloadEvicted)
			if c == nil {
				t.Fatal("expected WorkloadEvicted condition on new workload")
			}
			lat := now.Sub(c.LastTransitionTime.Time)
			if diff := cmp.Diff(tc.wantLatency, lat); diff != "" {
				t.Errorf("latency mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
