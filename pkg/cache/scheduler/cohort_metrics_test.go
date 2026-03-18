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

package scheduler

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"

	kueue "sigs.k8s.io/kueue/apis/kueue/v1beta2"
	kueuemetrics "sigs.k8s.io/kueue/pkg/metrics"
	"sigs.k8s.io/kueue/pkg/resources"
	utiltesting "sigs.k8s.io/kueue/pkg/util/testing"
	utiltestingmetrics "sigs.k8s.io/kueue/pkg/util/testing/metrics"
	utiltestingapi "sigs.k8s.io/kueue/pkg/util/testing/v1beta2"
)

func TestRecordCohortMetrics_Guards(t *testing.T) {
	type wantPoint struct {
		cohort       kueue.CohortReference
		fr           resources.FlavorResource
		quota        float64
		reservations float64
	}

	cohortLeft := kueue.CohortReference("record-left")
	cohortRight := kueue.CohortReference("record-right")
	cohortRoot := kueue.CohortReference("record-root")
	fr := resources.FlavorResource{Flavor: "flavor1", Resource: corev1.ResourceCPU}

	cases := []struct {
		name           string
		cohortToRecord kueue.CohortReference
		setup          func(t *testing.T, ctx context.Context, log logr.Logger, cache *Cache) []wantPoint
		wantNoMetrics  []kueue.CohortReference
	}{
		{
			name:           "Empty cohort name skips recording",
			cohortToRecord: "",
			setup:          func(t *testing.T, ctx context.Context, log logr.Logger, cache *Cache) []wantPoint { return nil },
			wantNoMetrics:  []kueue.CohortReference{"empty-sentinel"},
		},
		{
			name:           "Unknown cohort skips recording",
			cohortToRecord: "missing-cohort",
			setup:          func(t *testing.T, ctx context.Context, log logr.Logger, cache *Cache) []wantPoint { return nil },
			wantNoMetrics:  []kueue.CohortReference{"missing-cohort"},
		},
		{
			name:           "Records target cohort and ancestors only",
			cohortToRecord: "record-left",
			setup: func(t *testing.T, ctx context.Context, log logr.Logger, cache *Cache) []wantPoint {
				t.Helper()

				setupRecordMetricsHierarchy(
					ctx, t, log, cache,
					[]*kueue.ResourceFlavor{
						utiltestingapi.MakeResourceFlavor("flavor1").Obj(),
					},
					[]*kueue.Cohort{
						utiltestingapi.MakeCohort(cohortRoot).
							ResourceGroup(*utiltestingapi.MakeFlavorQuotas("flavor1").
								Resource(corev1.ResourceCPU, "30").
								Obj()).
							Obj(),
						utiltestingapi.MakeCohort(cohortLeft).
							Parent(cohortRoot).
							ResourceGroup(*utiltestingapi.MakeFlavorQuotas("flavor1").
								Resource(corev1.ResourceCPU, "20").
								Obj()).
							Obj(),
						utiltestingapi.MakeCohort(cohortRight).
							Parent(cohortRoot).
							ResourceGroup(*utiltestingapi.MakeFlavorQuotas("flavor1").
								Resource(corev1.ResourceCPU, "10").
								Obj()).
							Obj(),
					},
					[]*kueue.ClusterQueue{
						utiltestingapi.MakeClusterQueue("record-left-cq").
							Cohort(cohortLeft).
							ResourceGroup(*utiltestingapi.MakeFlavorQuotas("flavor1").
								Resource(corev1.ResourceCPU, "10").
								Obj()).
							Obj(),
						utiltestingapi.MakeClusterQueue("record-right-cq").
							Cohort(cohortRight).
							ResourceGroup(*utiltestingapi.MakeFlavorQuotas("flavor1").
								Resource(corev1.ResourceCPU, "5").
								Obj()).
							Obj(),
					},
				)

				return []wantPoint{
					{
						cohort: cohortLeft,
						fr:     fr,
						quota:  float64(cache.hm.Cohort(cohortLeft).resourceNode.SubtreeQuota[fr]),
					},
					{
						cohort: cohortRoot,
						fr:     fr,
						quota:  float64(cache.hm.Cohort(cohortRoot).resourceNode.SubtreeQuota[fr]),
					},
				}
			},
			wantNoMetrics: []kueue.CohortReference{cohortRight},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, log := utiltesting.ContextWithLog(t)
			cache := New(utiltesting.NewFakeClient())

			wantPoints := tc.setup(t, ctx, log, cache)

			// Ensure clean slate for all cohorts referenced in this test case.
			cohortsToClear := make([]kueue.CohortReference, 0, len(tc.wantNoMetrics)+len(wantPoints)+1)
			cohortsToClear = append(cohortsToClear, tc.wantNoMetrics...)
			cohortsToClear = append(cohortsToClear, tc.cohortToRecord)
			for _, p := range wantPoints {
				cohortsToClear = append(cohortsToClear, p.cohort)
			}
			clearCohortMetricsForTest(t, cohortsToClear...)

			cache.RecordCohortMetrics(log, tc.cohortToRecord)

			for _, want := range wantPoints {
				labels := cohortQuotaMetricLabels(want.cohort, want.fr)
				expectGaugeValue(t, kueuemetrics.CohortSubtreeQuota, labels, want.quota)
				if want.reservations != 0 {
					expectGaugeValue(t, kueuemetrics.CohortSubtreeResourceReservations, labels, want.reservations)
				} else {
					expectGaugeCount(t, kueuemetrics.CohortSubtreeResourceReservations, 0, labels)
				}
			}

			for _, cohortName := range tc.wantNoMetrics {
				expectGaugeCount(t, kueuemetrics.CohortSubtreeQuota, 0, map[string]string{"cohort": string(cohortName)})
				expectGaugeCount(t, kueuemetrics.CohortSubtreeResourceReservations, 0, map[string]string{"cohort": string(cohortName)})
			}
		})
	}
}

func TestRecordCohortMetrics_QuotaHierarchyLikeIntegration(t *testing.T) {
	fixture := newCohortMetricsFixture(t)
	cache := fixture.cache
	ctx, log := fixture.ctx, fixture.log

	frDefaultCPU := resources.FlavorResource{Flavor: "default", Resource: corev1.ResourceCPU}
	frFlavor1CPU := resources.FlavorResource{Flavor: "flavor1", Resource: corev1.ResourceCPU}
	frFlavor1GPU := resources.FlavorResource{Flavor: "flavor1", Resource: corev1.ResourceName("nvidia.com/gpu")}

	setupRecordMetricsHierarchy(ctx, t, log, cache,
		[]*kueue.ResourceFlavor{
			utiltestingapi.MakeResourceFlavor("default").Obj(),
			utiltestingapi.MakeResourceFlavor("flavor1").Obj(),
		},
		[]*kueue.Cohort{
			utiltestingapi.MakeCohort("root").
				ResourceGroup(*utiltestingapi.MakeFlavorQuotas("flavor1").
					Resource(corev1.ResourceCPU, "20").
					Resource("nvidia.com/gpu", "5").
					Obj()).
				Obj(),
			utiltestingapi.MakeCohort("ch1").
				Parent("root").
				ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
					Resource(corev1.ResourceCPU, "15").
					Obj()).
				Obj(),
			utiltestingapi.MakeCohort("ch2").
				Parent("root").
				ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
					Resource(corev1.ResourceCPU, "10").
					Obj()).
				Obj(),
			utiltestingapi.MakeCohort("ch3").
				Parent("root").
				ResourceGroup(*utiltestingapi.MakeFlavorQuotas("flavor1").
					Resource(corev1.ResourceCPU, "5").
					Resource("nvidia.com/gpu", "1").
					Obj()).
				Obj(),
		},
		[]*kueue.ClusterQueue{
			utiltestingapi.MakeClusterQueue("cqa").
				Cohort("ch1").
				ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").Resource(corev1.ResourceCPU, "10").Obj()).
				Obj(),
			utiltestingapi.MakeClusterQueue("cqb").
				Cohort("ch1").
				ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").Resource(corev1.ResourceCPU, "5").Obj()).
				Obj(),
			utiltestingapi.MakeClusterQueue("cqd").
				Cohort("ch2").
				ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").Resource(corev1.ResourceCPU, "5").Obj()).
				Obj(),
			utiltestingapi.MakeClusterQueue("cqe").
				Cohort("ch2").
				ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").Resource(corev1.ResourceCPU, "5").Obj()).
				Obj(),
		},
	)
	t.Run("QuotaHierarchyLikeIntegration", func(t *testing.T) {
		clearCohortMetricsForTest(t, "root", "ch1", "ch2", "ch3")

		cache.RecordCohortMetrics(log, "ch1")
		cache.RecordCohortMetrics(log, "ch2")
		cache.RecordCohortMetrics(log, "ch3")

		expectGaugeValue(t, kueuemetrics.CohortSubtreeQuota, cohortQuotaMetricLabels("ch1", frDefaultCPU), 30_000)
		expectGaugeValue(t, kueuemetrics.CohortSubtreeQuota, cohortQuotaMetricLabels("ch2", frDefaultCPU), 20_000)
		expectGaugeValue(t, kueuemetrics.CohortSubtreeQuota, cohortQuotaMetricLabels("ch3", frFlavor1CPU), 5_000)
		expectGaugeValue(t, kueuemetrics.CohortSubtreeQuota, cohortQuotaMetricLabels("ch3", frFlavor1GPU), 1)

		expectGaugeValue(t, kueuemetrics.CohortSubtreeQuota, cohortQuotaMetricLabels("root", frDefaultCPU), 50_000)
		expectGaugeValue(t, kueuemetrics.CohortSubtreeQuota, cohortQuotaMetricLabels("root", frFlavor1CPU), 25_000)
		expectGaugeValue(t, kueuemetrics.CohortSubtreeQuota, cohortQuotaMetricLabels("root", frFlavor1GPU), 6)
	})
}

func TestRecordCohortMetrics_ReservationsChildParentLikeIntegration(t *testing.T) {
	fixture := newCohortMetricsFixture(t)
	cache := fixture.cache
	ctx, log := fixture.ctx, fixture.log
	frDefaultCPU := resources.FlavorResource{Flavor: "default", Resource: corev1.ResourceCPU}

	setupRecordMetricsHierarchy(
		ctx, t, log, cache,
		[]*kueue.ResourceFlavor{
			utiltestingapi.MakeResourceFlavor("default").Obj(),
		},
		[]*kueue.Cohort{
			utiltestingapi.MakeCohort("root").
				ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").Resource(corev1.ResourceCPU, "30").Obj()).
				Obj(),
			utiltestingapi.MakeCohort("ch1").
				Parent("root").
				ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").Resource(corev1.ResourceCPU, "8").Obj()).
				Obj(),
			utiltestingapi.MakeCohort("ch2").
				Parent("root").
				ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").Resource(corev1.ResourceCPU, "6").Obj()).
				Obj(),
		},
		[]*kueue.ClusterQueue{
			utiltestingapi.MakeClusterQueue("cqa").
				Cohort("ch1").
				ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").Resource(corev1.ResourceCPU, "10").Obj()).
				Obj(),
			utiltestingapi.MakeClusterQueue("cqb").
				Cohort("ch2").
				ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").Resource(corev1.ResourceCPU, "8").Obj()).
				Obj(),
		},
	)

	clearCohortMetricsForTest(t, "root", "ch1", "ch2")

	addTestUsage(cache, []usageChange{
		{cqName: "cqa", fr: frDefaultCPU, val: 6_000},
	})
	cache.RecordCohortMetrics(log, "ch1")
	expectGaugeValue(t, kueuemetrics.CohortSubtreeResourceReservations, cohortQuotaMetricLabels("ch1", frDefaultCPU), 6_000)
	expectGaugeValue(t, kueuemetrics.CohortSubtreeResourceReservations, cohortQuotaMetricLabels("root", frDefaultCPU), 6_000)

	addTestUsage(cache, []usageChange{
		{cqName: "cqa", fr: frDefaultCPU, val: 16_000},
	})
	cache.RecordCohortMetrics(log, "ch1")
	expectGaugeValue(t, kueuemetrics.CohortSubtreeResourceReservations, cohortQuotaMetricLabels("ch1", frDefaultCPU), 16_000)
	expectGaugeValue(t, kueuemetrics.CohortSubtreeResourceReservations, cohortQuotaMetricLabels("root", frDefaultCPU), 16_000)

	addTestUsage(cache, []usageChange{
		{cqName: "cqb", fr: frDefaultCPU, val: 7_000},
	})
	cache.RecordCohortMetrics(log, "ch2")
	expectGaugeValue(t, kueuemetrics.CohortSubtreeResourceReservations, cohortQuotaMetricLabels("ch2", frDefaultCPU), 7_000)
	expectGaugeValue(t, kueuemetrics.CohortSubtreeResourceReservations, cohortQuotaMetricLabels("root", frDefaultCPU), 23_000)

	addTestUsage(cache, []usageChange{
		{cqName: "cqa", fr: frDefaultCPU, val: 6_000},
	})
	cache.RecordCohortMetrics(log, "ch1")
	expectGaugeValue(t, kueuemetrics.CohortSubtreeResourceReservations, cohortQuotaMetricLabels("ch1", frDefaultCPU), 6_000)
	expectGaugeValue(t, kueuemetrics.CohortSubtreeResourceReservations, cohortQuotaMetricLabels("ch2", frDefaultCPU), 7_000)
	expectGaugeValue(t, kueuemetrics.CohortSubtreeResourceReservations, cohortQuotaMetricLabels("root", frDefaultCPU), 13_000)
}

func TestClearCohortMetrics_ClearsTargetAndAncestors(t *testing.T) {
	fixture := newCohortMetricsFixture(t)
	cache := fixture.cache
	ctx, log := fixture.ctx, fixture.log
	frDefaultCPU := resources.FlavorResource{Flavor: "default", Resource: corev1.ResourceCPU}

	setupRecordMetricsHierarchy(
		ctx, t, log, cache,
		[]*kueue.ResourceFlavor{
			utiltestingapi.MakeResourceFlavor("default").Obj(),
		},
		[]*kueue.Cohort{
			utiltestingapi.MakeCohort("root").
				ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").Resource(corev1.ResourceCPU, "30").Obj()).
				Obj(),
			utiltestingapi.MakeCohort("ch1").
				Parent("root").
				ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").Resource(corev1.ResourceCPU, "8").Obj()).
				Obj(),
			utiltestingapi.MakeCohort("ch2").
				Parent("root").
				ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").Resource(corev1.ResourceCPU, "6").Obj()).
				Obj(),
		},
		[]*kueue.ClusterQueue{
			utiltestingapi.MakeClusterQueue("cqa").
				Cohort("ch1").
				ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").Resource(corev1.ResourceCPU, "10").Obj()).
				Obj(),
			utiltestingapi.MakeClusterQueue("cqb").
				Cohort("ch2").
				ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").Resource(corev1.ResourceCPU, "8").Obj()).
				Obj(),
		},
	)

	clearCohortMetricsForTest(t, "root", "ch1", "ch2")
	addTestUsage(cache, []usageChange{
		{cqName: "cqa", fr: frDefaultCPU, val: 6_000},
		{cqName: "cqb", fr: frDefaultCPU, val: 7_000},
	})

	cache.RecordCohortMetrics(log, "ch1")
	cache.RecordCohortMetrics(log, "ch2")
	expectGaugeValue(t, kueuemetrics.CohortSubtreeQuota, cohortQuotaMetricLabels("ch1", frDefaultCPU), 18_000)
	expectGaugeValue(t, kueuemetrics.CohortSubtreeQuota, cohortQuotaMetricLabels("ch2", frDefaultCPU), 14_000)
	expectGaugeValue(t, kueuemetrics.CohortSubtreeQuota, cohortQuotaMetricLabels("root", frDefaultCPU), 62_000)
	expectGaugeValue(t, kueuemetrics.CohortSubtreeResourceReservations, cohortQuotaMetricLabels("root", frDefaultCPU), 13_000)

	cache.ClearCohortMetrics(log, "ch1")
	expectGaugeCount(t, kueuemetrics.CohortSubtreeQuota, 0, cohortQuotaMetricLabels("ch1", frDefaultCPU))
	expectGaugeCount(t, kueuemetrics.CohortSubtreeResourceReservations, 0, cohortQuotaMetricLabels("ch1", frDefaultCPU))
	expectGaugeValue(t, kueuemetrics.CohortSubtreeQuota, cohortQuotaMetricLabels("ch2", frDefaultCPU), 14_000)
	expectGaugeValue(t, kueuemetrics.CohortSubtreeResourceReservations, cohortQuotaMetricLabels("ch2", frDefaultCPU), 7_000)
	expectGaugeValue(t, kueuemetrics.CohortSubtreeQuota, cohortQuotaMetricLabels("root", frDefaultCPU), 44_000)
	expectGaugeValue(t, kueuemetrics.CohortSubtreeResourceReservations, cohortQuotaMetricLabels("root", frDefaultCPU), 7_000)
}

func TestRecordCohortSubtreeInfoMetrics(t *testing.T) {
	t.Run("unknown cohort is a no-op", func(t *testing.T) {
		fixture := newCohortMetricsFixture(t)
		cache := fixture.cache

		clearCohortMetricsForTest(t, "unknown")
		cache.RecordCohortSubtreeInfoMetrics("unknown")

		expectGaugeCount(t, kueuemetrics.CohortInfo, 0, map[string]string{"cohort": "unknown"})
	})

	t.Run("single root cohort reports self with empty parent", func(t *testing.T) {
		fixture := newCohortMetricsFixture(t)
		cache := fixture.cache
		ctx, log := fixture.ctx, fixture.log

		setupRecordMetricsHierarchy(ctx, t, log, cache,
			[]*kueue.ResourceFlavor{
				utiltestingapi.MakeResourceFlavor("default").Obj(),
			},
			[]*kueue.Cohort{
				utiltestingapi.MakeCohort("root").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "10").Obj()).
					Obj(),
			},
			[]*kueue.ClusterQueue{},
		)
		clearCohortMetricsForTest(t, "root")
		cache.RecordCohortSubtreeInfoMetrics("root")

		expectGaugeValue(t, kueuemetrics.CohortInfo, cohortMetricInfoLabels("root", "", "root"), 1)
	})

	t.Run("root cohort with child CQs reports cohort and CQ info", func(t *testing.T) {
		fixture := newCohortMetricsFixture(t)
		cache := fixture.cache
		ctx, log := fixture.ctx, fixture.log

		setupRecordMetricsHierarchy(ctx, t, log, cache,
			[]*kueue.ResourceFlavor{
				utiltestingapi.MakeResourceFlavor("default").Obj(),
			},
			[]*kueue.Cohort{
				utiltestingapi.MakeCohort("root").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "10").Obj()).
					Obj(),
			},
			[]*kueue.ClusterQueue{
				utiltestingapi.MakeClusterQueue("cq1").
					Cohort("root").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "5").Obj()).
					Obj(),
				utiltestingapi.MakeClusterQueue("cq2").
					Cohort("root").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "5").Obj()).
					Obj(),
			},
		)
		clearCohortMetricsForTest(t, "root")
		cache.RecordCohortSubtreeInfoMetrics("root")

		expectGaugeValue(t, kueuemetrics.CohortInfo, cohortMetricInfoLabels("root", "", "root"), 1)
		expectGaugeValue(t, kueuemetrics.ClusterQueueInfo, cqMetricInfoLabels("cq1", "root", "root"), 1)
		expectGaugeValue(t, kueuemetrics.ClusterQueueInfo, cqMetricInfoLabels("cq2", "root", "root"), 1)
	})

	t.Run("parent-child cohorts called from parent reports both", func(t *testing.T) {
		fixture := newCohortMetricsFixture(t)
		cache := fixture.cache
		ctx, log := fixture.ctx, fixture.log

		setupRecordMetricsHierarchy(ctx, t, log, cache,
			[]*kueue.ResourceFlavor{
				utiltestingapi.MakeResourceFlavor("default").Obj(),
			},
			[]*kueue.Cohort{
				utiltestingapi.MakeCohort("parent").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "10").Obj()).
					Obj(),
				utiltestingapi.MakeCohort("child").
					Parent("parent").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "5").Obj()).
					Obj(),
			},
			[]*kueue.ClusterQueue{},
		)
		clearCohortMetricsForTest(t, "parent", "child")
		cache.RecordCohortSubtreeInfoMetrics("parent")

		expectGaugeValue(t, kueuemetrics.CohortInfo, cohortMetricInfoLabels("parent", "", "parent"), 1)
		expectGaugeValue(t, kueuemetrics.CohortInfo, cohortMetricInfoLabels("child", "parent", "parent"), 1)
	})

	t.Run("parent-child cohorts called from child reports only child subtree", func(t *testing.T) {
		fixture := newCohortMetricsFixture(t)
		cache := fixture.cache
		ctx, log := fixture.ctx, fixture.log

		setupRecordMetricsHierarchy(ctx, t, log, cache,
			[]*kueue.ResourceFlavor{
				utiltestingapi.MakeResourceFlavor("default").Obj(),
			},
			[]*kueue.Cohort{
				utiltestingapi.MakeCohort("parent").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "10").Obj()).
					Obj(),
				utiltestingapi.MakeCohort("child").
					Parent("parent").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "5").Obj()).
					Obj(),
			},
			[]*kueue.ClusterQueue{},
		)
		clearCohortMetricsForTest(t, "parent", "child")
		cache.RecordCohortSubtreeInfoMetrics("child")

		expectGaugeValue(t, kueuemetrics.CohortInfo, cohortMetricInfoLabels("child", "parent", "parent"), 1)
		// Parent should not have metrics reported when called from child
		expectGaugeCount(t, kueuemetrics.CohortInfo, 0, cohortMetricInfoLabels("parent", "", "parent"))
	})

	t.Run("three-level hierarchy from root reports all descendants", func(t *testing.T) {
		fixture := newCohortMetricsFixture(t)
		cache := fixture.cache
		ctx, log := fixture.ctx, fixture.log

		setupRecordMetricsHierarchy(ctx, t, log, cache,
			[]*kueue.ResourceFlavor{
				utiltestingapi.MakeResourceFlavor("default").Obj(),
			},
			[]*kueue.Cohort{
				utiltestingapi.MakeCohort("root").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "10").Obj()).
					Obj(),
				utiltestingapi.MakeCohort("mid").
					Parent("root").Obj(),
				utiltestingapi.MakeCohort("leaf").
					Parent("mid").Obj(),
			},
			[]*kueue.ClusterQueue{
				utiltestingapi.MakeClusterQueue("cq-leaf").
					Cohort("leaf").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "1").Obj()).
					Obj(),
			},
		)
		clearCohortMetricsForTest(t, "root", "mid", "leaf")
		cache.RecordCohortSubtreeInfoMetrics("root")

		expectGaugeValue(t, kueuemetrics.CohortInfo, cohortMetricInfoLabels("root", "", "root"), 1)
		expectGaugeValue(t, kueuemetrics.CohortInfo, cohortMetricInfoLabels("mid", "root", "root"), 1)
		expectGaugeValue(t, kueuemetrics.CohortInfo, cohortMetricInfoLabels("leaf", "mid", "root"), 1)
		expectGaugeValue(t, kueuemetrics.ClusterQueueInfo, cqMetricInfoLabels("cq-leaf", "leaf", "root"), 1)
	})

	t.Run("three-level hierarchy from middle reports only middle and below", func(t *testing.T) {
		fixture := newCohortMetricsFixture(t)
		cache := fixture.cache
		ctx, log := fixture.ctx, fixture.log

		setupRecordMetricsHierarchy(ctx, t, log, cache,
			[]*kueue.ResourceFlavor{
				utiltestingapi.MakeResourceFlavor("default").Obj(),
			},
			[]*kueue.Cohort{
				utiltestingapi.MakeCohort("root").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "10").Obj()).
					Obj(),
				utiltestingapi.MakeCohort("mid").
					Parent("root").Obj(),
				utiltestingapi.MakeCohort("leaf").
					Parent("mid").Obj(),
			},
			[]*kueue.ClusterQueue{
				utiltestingapi.MakeClusterQueue("cq-leaf").
					Cohort("leaf").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "1").Obj()).
					Obj(),
			},
		)
		clearCohortMetricsForTest(t, "root", "mid", "leaf")
		cache.RecordCohortSubtreeInfoMetrics("mid")

		// Root not reported from mid
		expectGaugeCount(t, kueuemetrics.CohortInfo, 0, cohortMetricInfoLabels("root", "", "root"))
		// mid and leaf reported, both with root as root_cohort
		expectGaugeValue(t, kueuemetrics.CohortInfo, cohortMetricInfoLabels("mid", "root", "root"), 1)
		expectGaugeValue(t, kueuemetrics.CohortInfo, cohortMetricInfoLabels("leaf", "mid", "root"), 1)
		expectGaugeValue(t, kueuemetrics.ClusterQueueInfo, cqMetricInfoLabels("cq-leaf", "leaf", "root"), 1)
	})

	t.Run("sibling isolation: calling on one sibling does not report the other", func(t *testing.T) {
		fixture := newCohortMetricsFixture(t)
		cache := fixture.cache
		ctx, log := fixture.ctx, fixture.log

		setupRecordMetricsHierarchy(ctx, t, log, cache,
			[]*kueue.ResourceFlavor{
				utiltestingapi.MakeResourceFlavor("default").Obj(),
			},
			[]*kueue.Cohort{
				utiltestingapi.MakeCohort("root").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "10").Obj()).
					Obj(),
				utiltestingapi.MakeCohort("left").
					Parent("root").Obj(),
				utiltestingapi.MakeCohort("right").
					Parent("root").Obj(),
			},
			[]*kueue.ClusterQueue{
				utiltestingapi.MakeClusterQueue("cq-left").
					Cohort("left").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "2").Obj()).
					Obj(),
				utiltestingapi.MakeClusterQueue("cq-right").
					Cohort("right").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "3").Obj()).
					Obj(),
			},
		)
		clearCohortMetricsForTest(t, "root", "left", "right")
		cache.RecordCohortSubtreeInfoMetrics("left")

		expectGaugeValue(t, kueuemetrics.CohortInfo, cohortMetricInfoLabels("left", "root", "root"), 1)
		expectGaugeValue(t, kueuemetrics.ClusterQueueInfo, cqMetricInfoLabels("cq-left", "left", "root"), 1)
		// right sibling and its CQ should not have metrics
		expectGaugeCount(t, kueuemetrics.CohortInfo, 0, cohortMetricInfoLabels("right", "root", "root"))
		expectGaugeCount(t, kueuemetrics.ClusterQueueInfo, 0, cqMetricInfoLabels("cq-right", "right", "root"))
	})

	t.Run("implicit parent: calling on implicit parent name reports parent and child", func(t *testing.T) {
		fixture := newCohortMetricsFixture(t)
		cache := fixture.cache
		ctx, log := fixture.ctx, fixture.log

		setupRecordMetricsHierarchy(ctx, t, log, cache,
			[]*kueue.ResourceFlavor{
				utiltestingapi.MakeResourceFlavor("default").Obj(),
			},
			[]*kueue.Cohort{
				utiltestingapi.MakeCohort("child").
					Parent("implicit-parent").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "10").Obj()).
					Obj(),
			},
			[]*kueue.ClusterQueue{},
		)
		if cache.hm.Cohort("implicit-parent") == nil {
			t.Fatal("implicit parent cohort was not created")
		}

		clearCohortMetricsForTest(t, "child", "implicit-parent")
		cache.RecordCohortSubtreeInfoMetrics("implicit-parent")

		expectGaugeValue(t, kueuemetrics.CohortInfo, cohortMetricInfoLabels("implicit-parent", "", "implicit-parent"), 1)
		expectGaugeValue(t, kueuemetrics.CohortInfo, cohortMetricInfoLabels("child", "implicit-parent", "implicit-parent"), 1)
	})

	t.Run("implicit parent: calling on child does not report implicit parent", func(t *testing.T) {
		fixture := newCohortMetricsFixture(t)
		cache := fixture.cache
		ctx, log := fixture.ctx, fixture.log

		setupRecordMetricsHierarchy(ctx, t, log, cache,
			[]*kueue.ResourceFlavor{
				utiltestingapi.MakeResourceFlavor("default").Obj(),
			},
			[]*kueue.Cohort{
				utiltestingapi.MakeCohort("child").
					Parent("implicit-parent").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "10").Obj()).
					Obj(),
			},
			[]*kueue.ClusterQueue{},
		)
		clearCohortMetricsForTest(t, "child", "implicit-parent")
		cache.RecordCohortSubtreeInfoMetrics("child")

		expectGaugeValue(t, kueuemetrics.CohortInfo, cohortMetricInfoLabels("child", "implicit-parent", "implicit-parent"), 1)
		expectGaugeCount(t, kueuemetrics.CohortInfo, 0, cohortMetricInfoLabels("implicit-parent", "", "implicit-parent"))
	})

	t.Run("implicit parent with multiple children reports all children", func(t *testing.T) {
		fixture := newCohortMetricsFixture(t)
		cache := fixture.cache
		ctx, log := fixture.ctx, fixture.log

		setupRecordMetricsHierarchy(ctx, t, log, cache,
			[]*kueue.ResourceFlavor{
				utiltestingapi.MakeResourceFlavor("default").Obj(),
			},
			[]*kueue.Cohort{
				utiltestingapi.MakeCohort("child-a").
					Parent("implicit-root").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "5").Obj()).
					Obj(),
				utiltestingapi.MakeCohort("child-b").
					Parent("implicit-root").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "5").Obj()).
					Obj(),
			},
			[]*kueue.ClusterQueue{
				utiltestingapi.MakeClusterQueue("cq-a").
					Cohort("child-a").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "2").Obj()).
					Obj(),
			},
		)
		clearCohortMetricsForTest(t, "implicit-root", "child-a", "child-b")
		cache.RecordCohortSubtreeInfoMetrics("implicit-root")

		expectGaugeValue(t, kueuemetrics.CohortInfo, cohortMetricInfoLabels("implicit-root", "", "implicit-root"), 1)
		expectGaugeValue(t, kueuemetrics.CohortInfo, cohortMetricInfoLabels("child-a", "implicit-root", "implicit-root"), 1)
		expectGaugeValue(t, kueuemetrics.CohortInfo, cohortMetricInfoLabels("child-b", "implicit-root", "implicit-root"), 1)
		expectGaugeValue(t, kueuemetrics.ClusterQueueInfo, cqMetricInfoLabels("cq-a", "child-a", "implicit-root"), 1)
	})

	t.Run("clears stale series before reporting updated hierarchy", func(t *testing.T) {
		fixture := newCohortMetricsFixture(t)
		cache := fixture.cache
		ctx, log := fixture.ctx, fixture.log

		setupRecordMetricsHierarchy(ctx, t, log, cache,
			[]*kueue.ResourceFlavor{
				utiltestingapi.MakeResourceFlavor("default").Obj(),
			},
			[]*kueue.Cohort{
				utiltestingapi.MakeCohort("root").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "10").Obj()).
					Obj(),
				utiltestingapi.MakeCohort("child").
					Parent("root").Obj(),
			},
			[]*kueue.ClusterQueue{
				utiltestingapi.MakeClusterQueue("cq1").
					Cohort("child").
					ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
						Resource(corev1.ResourceCPU, "5").Obj()).
					Obj(),
			},
		)
		clearCohortMetricsForTest(t, "root", "child", "new-root")

		// First report: child under root
		cache.RecordCohortSubtreeInfoMetrics("root")
		expectGaugeValue(t, kueuemetrics.CohortInfo, cohortMetricInfoLabels("child", "root", "root"), 1)
		expectGaugeValue(t, kueuemetrics.ClusterQueueInfo, cqMetricInfoLabels("cq1", "child", "root"), 1)

		// Reparent child under a new root
		newRoot := utiltestingapi.MakeCohort("new-root").
			ResourceGroup(*utiltestingapi.MakeFlavorQuotas("default").
				Resource(corev1.ResourceCPU, "20").Obj()).
			Obj()
		if _, err := cache.AddOrUpdateCohort(newRoot); err != nil {
			t.Fatalf("adding new-root: %v", err)
		}
		childCohort := utiltestingapi.MakeCohort("child").
			Parent("new-root").Obj()
		if _, err := cache.AddOrUpdateCohort(childCohort); err != nil {
			t.Fatalf("reparenting child: %v", err)
		}

		// Report from new-root: stale root labels should be replaced
		cache.RecordCohortSubtreeInfoMetrics("new-root")
		expectGaugeValue(t, kueuemetrics.CohortInfo, cohortMetricInfoLabels("child", "new-root", "new-root"), 1)
		expectGaugeValue(t, kueuemetrics.ClusterQueueInfo, cqMetricInfoLabels("cq1", "child", "new-root"), 1)
		expectGaugeValue(t, kueuemetrics.CohortInfo, cohortMetricInfoLabels("new-root", "", "new-root"), 1)
	})
}

type cohortMetricsFixture struct {
	ctx   context.Context
	log   logr.Logger
	cache *Cache
}

func newCohortMetricsFixture(t *testing.T) cohortMetricsFixture {
	t.Helper()
	ctx, log := utiltesting.ContextWithLog(t)
	return cohortMetricsFixture{ctx: ctx, log: log, cache: New(utiltesting.NewFakeClient())}
}

func setupRecordMetricsHierarchy(
	ctx context.Context,
	t *testing.T,
	log logr.Logger,
	cache *Cache,
	flavors []*kueue.ResourceFlavor,
	cohorts []*kueue.Cohort,
	cqs []*kueue.ClusterQueue,
) {
	t.Helper()
	for _, flavor := range flavors {
		cache.AddOrUpdateResourceFlavor(log, flavor)
	}

	for _, ch := range cohorts {
		if err := cache.AddOrUpdateCohort(ch); err != nil {
			t.Fatalf("adding cohort %q: %v", ch.Name, err)
		}
	}

	for _, cq := range cqs {
		if err := cache.AddClusterQueue(ctx, cq); err != nil {
			t.Fatalf("adding clusterQueue %q: %v", cq.Name, err)
		}
	}
}

type usageChange struct {
	cqName kueue.ClusterQueueReference
	fr     resources.FlavorResource
	val    int64
}

// addUsage adds usage to the current node, and adds usage past localQuota
// to its Cohort.
func addTestUsage(cache *Cache, fr []usageChange) {
	for _, val := range fr {
		cache.hm.ClusterQueue(val.cqName).resourceNode.Usage[val.fr] = val.val
	}
}

func clearCohortMetricsForTest(t *testing.T, cohorts ...kueue.CohortReference) {
	t.Helper()
	for _, cohort := range cohorts {
		if cohort != "" {
			kueuemetrics.ClearCohortMetrics(cohort)
			kueuemetrics.ClearCohortInfo(cohort)
		}
	}
	t.Cleanup(func() {
		for _, cohort := range cohorts {
			if cohort != "" {
				kueuemetrics.ClearCohortMetrics(cohort)
				kueuemetrics.ClearCohortInfo(cohort)
			}
		}
	})
}

func cohortQuotaMetricLabels(cohortName kueue.CohortReference, fr resources.FlavorResource) map[string]string {
	return map[string]string{
		"cohort":       string(cohortName),
		"flavor":       string(fr.Flavor),
		"resource":     string(fr.Resource),
		"replica_role": "standalone",
	}
}

func cohortMetricInfoLabels(cohortName kueue.CohortReference, parentCohort, rootCohort string) map[string]string {
	return map[string]string{
		"cohort":        string(cohortName),
		"parent_cohort": parentCohort,
		"root_cohort":   rootCohort,
		"replica_role":  "standalone",
	}
}

func cqMetricInfoLabels(cqName kueue.ClusterQueueReference, parentCohort, rootCohort string) map[string]string {
	return map[string]string{
		"cluster_queue": string(cqName),
		"parent_cohort": parentCohort,
		"root_cohort":   rootCohort,
		"replica_role":  "standalone",
	}
}

func expectGaugeCount(t *testing.T, collector prometheus.Collector, want int, labels map[string]string) {
	t.Helper()
	got := len(utiltestingmetrics.CollectFilteredGaugeVec(collector, labels))
	if got != want {
		t.Fatalf("unexpected metric count for labels %v: got=%d want=%d", labels, got, want)
	}
}

func expectGaugeValue(t *testing.T, collector prometheus.Collector, labels map[string]string, want float64) {
	t.Helper()
	dps := utiltestingmetrics.CollectFilteredGaugeVec(collector, labels)
	if len(dps) != 1 {
		t.Fatalf("expected exactly one metric for labels %v, got=%d", labels, len(dps))
	}
	if dps[0].Value != want {
		t.Fatalf("unexpected metric value for labels %v: got=%v want=%v", labels, dps[0].Value, want)
	}
}
