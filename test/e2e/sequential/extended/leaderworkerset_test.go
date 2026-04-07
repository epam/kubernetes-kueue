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

package extended

import (
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	leaderworkersetv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"

	config "sigs.k8s.io/kueue/apis/config/v1beta2"
	kueue "sigs.k8s.io/kueue/apis/kueue/v1beta2"
	"sigs.k8s.io/kueue/pkg/features"
	utiltestingapi "sigs.k8s.io/kueue/pkg/util/testing/v1beta2"
	leaderworkersettesting "sigs.k8s.io/kueue/pkg/util/testingjobs/leaderworkerset"
	"sigs.k8s.io/kueue/test/util"
)

var _ = ginkgo.Describe("LeaderWorkerSet WorkloadIdentifierAnnotations", ginkgo.Ordered, ginkgo.ContinueOnFailure, ginkgo.Label("feature:leaderworkerset"), func() {
	var (
		ns *corev1.Namespace
		rf *kueue.ResourceFlavor
		cq *kueue.ClusterQueue
		lq *kueue.LocalQueue
	)

	ginkgo.BeforeEach(func() {
		ns = util.CreateNamespaceFromPrefixWithLog(ctx, k8sClient, "lws-e2e-")

		rf = utiltestingapi.MakeResourceFlavor("rf-"+ns.Name).NodeLabel("instance-type", "on-demand").Obj()
		util.MustCreate(ctx, k8sClient, rf)

		cq = utiltestingapi.MakeClusterQueue("cq-" + ns.Name).
			ResourceGroup(
				*utiltestingapi.MakeFlavorQuotas(rf.Name).
					Resource(corev1.ResourceCPU, "5").
					Obj(),
			).
			Obj()
		util.CreateClusterQueuesAndWaitForActive(ctx, k8sClient, cq)

		lq = utiltestingapi.MakeLocalQueue("lq-"+ns.Name, ns.Name).ClusterQueue(cq.Name).Obj()
		util.CreateLocalQueuesAndWaitForActive(ctx, k8sClient, lq)
	})

	ginkgo.AfterEach(func() {
		gomega.Expect(util.DeleteAllLeaderWorkerSetsInNamespace(ctx, k8sClient, ns)).To(gomega.Succeed())
		gomega.Expect(util.DeleteNamespace(ctx, k8sClient, ns)).To(gomega.Succeed())
		util.ExpectObjectToBeDeleted(ctx, k8sClient, cq, true)
		util.ExpectObjectToBeDeleted(ctx, k8sClient, rf, true)
		util.ExpectAllPodsInNamespaceDeleted(ctx, k8sClient, ns)
	})

	ginkgo.Context("with WorkloadIdentifierAnnotations disabled", func() {
		ginkgo.BeforeAll(func() {
			if defaultKueueCfg.FeatureGates[string(features.WorkloadIdentifierAnnotations)] != false {
				util.UpdateKueueConfigurationAndRestart(ctx, k8sClient, defaultKueueCfg, kindClusterName, func(cfg *config.Configuration) {
					if cfg.FeatureGates == nil {
						cfg.FeatureGates = make(map[string]bool)
					}
					cfg.FeatureGates[string(features.WorkloadIdentifierAnnotations)] = false
				})
			}
		})

		ginkgo.It("should admit group with 39-character lws name (result name is 63 chars)", func() {
			lwsName := strings.Repeat("a", 39)
			lws := leaderworkersettesting.MakeLeaderWorkerSet(lwsName, ns.Name).
				Image(util.GetAgnHostImage(), util.BehaviorWaitForDeletion).
				Size(1).
				Replicas(1).
				RequestAndLimit(corev1.ResourceCPU, "200m").
				TerminationGracePeriod(1).
				Queue(lq.Name).
				Obj()

			ginkgo.By("create a LeaderWorkerSet", func() {
				util.MustCreate(ctx, k8sClient, lws)
			})

			ginkgo.By("waiting for replicas to be ready", func() {
				createdLeaderWorkerSet := &leaderworkersetv1.LeaderWorkerSet{}
				gomega.Eventually(func(g gomega.Gomega) {
					g.Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(lws), createdLeaderWorkerSet)).To(gomega.Succeed())
					g.Expect(createdLeaderWorkerSet.Status.ReadyReplicas).To(gomega.Equal(int32(1)))
				}, util.LongTimeout, util.Interval).Should(gomega.Succeed())
			})
		})

		ginkgo.It("should fail to admit group with 40-character lws name (kueue workload name exceeds label limit)", func() {
			lwsName := strings.Repeat("a", 40)
			lws := leaderworkersettesting.MakeLeaderWorkerSet(lwsName, ns.Name).
				Image(util.GetAgnHostImage(), util.BehaviorWaitForDeletion).
				Size(1).Replicas(1).Queue(lq.Name).Obj()

			ginkgo.By("create a LeaderWorkerSet", func() {
				util.MustCreate(ctx, k8sClient, lws)
			})

			ginkgo.By("confirming no replicas become ready", func() {
				createdLeaderWorkerSet := &leaderworkersetv1.LeaderWorkerSet{}
				gomega.Consistently(func(g gomega.Gomega) {
					g.Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(lws), createdLeaderWorkerSet)).To(gomega.Succeed())
					g.Expect(createdLeaderWorkerSet.Status.ReadyReplicas).To(gomega.Equal(int32(0)))
				}, util.ConsistentDuration, util.Interval).Should(gomega.Succeed())
			})
		})
	})

	ginkgo.Context("with WorkloadIdentifierAnnotations enabled", func() {
		ginkgo.BeforeAll(func() {
			util.UpdateKueueConfigurationAndRestart(ctx, k8sClient, defaultKueueCfg, kindClusterName, func(cfg *config.Configuration) {
				if cfg.FeatureGates == nil {
					cfg.FeatureGates = make(map[string]bool)
				}
				cfg.FeatureGates[string(features.WorkloadIdentifierAnnotations)] = true
			})
		})

		ginkgo.It("should admit group with 53-character lws name", func() {
			lwsName := strings.Repeat("a", 53)
			lws := leaderworkersettesting.MakeLeaderWorkerSet(lwsName, ns.Name).
				Image(util.GetAgnHostImage(), util.BehaviorWaitForDeletion).
				Size(1).Replicas(1).
				RequestAndLimit(corev1.ResourceCPU, "200m").
				TerminationGracePeriod(1).
				Queue(lq.Name).Obj()

			ginkgo.By("create a LeaderWorkerSet", func() {
				util.MustCreate(ctx, k8sClient, lws)
			})

			ginkgo.By("waiting for replicas to be ready", func() {
				createdLeaderWorkerSet := &leaderworkersetv1.LeaderWorkerSet{}
				gomega.Eventually(func(g gomega.Gomega) {
					g.Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(lws), createdLeaderWorkerSet)).To(gomega.Succeed())
					g.Expect(createdLeaderWorkerSet.Status.ReadyReplicas).To(gomega.Equal(int32(1)))
				}, util.LongTimeout, util.Interval).Should(gomega.Succeed())
			})
		})

		ginkgo.It("should fail to admit group with 54-character lws name", func() {
			lwsName := strings.Repeat("a", 54)
			lws := leaderworkersettesting.MakeLeaderWorkerSet(lwsName, ns.Name).
				Image(util.GetAgnHostImage(), util.BehaviorWaitForDeletion).
				Size(1).Replicas(1).Queue(lq.Name).Obj()

			ginkgo.By("create a LeaderWorkerSet", func() {
				util.MustCreate(ctx, k8sClient, lws)
			})

			ginkgo.By("waiting for FailedCreate event", func() {
				gomega.Eventually(func(g gomega.Gomega) {
					eventList := &corev1.EventList{}
					g.Expect(k8sClient.List(ctx, eventList, client.InNamespace(ns.Name))).To(gomega.Succeed())
					var found bool
					for _, e := range eventList.Items {
						if e.Reason == "FailedCreate" && strings.Contains(e.Message, "must be no more than 63") {
							found = true
						}
					}
					g.Expect(found).To(gomega.BeTrue())
				}, util.LongTimeout, util.Interval).Should(gomega.Succeed())
			})

			ginkgo.By("confirming no replicas become ready", func() {
				createdLeaderWorkerSet := &leaderworkersetv1.LeaderWorkerSet{}
				gomega.Consistently(func(g gomega.Gomega) {
					g.Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(lws), createdLeaderWorkerSet)).To(gomega.Succeed())
					g.Expect(createdLeaderWorkerSet.Status.ReadyReplicas).To(gomega.Equal(int32(0)))
				}, util.ConsistentDuration, util.Interval).Should(gomega.Succeed())
			})
		})
	})
})
