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

package hierarchy

import (
	kueue "sigs.k8s.io/kueue/apis/kueue/v1beta2"
)

// InfoPoint carries hierarchy relationships for either a cohort or a
// direct child clusterQueue entry.
type InfoPoint struct {
	CohortName       *kueue.CohortReference
	ClusterQueueName *kueue.ClusterQueueReference
	ParentCohortName kueue.CohortReference
	RootCohortName   kueue.CohortReference
}

// CohortInfoPoints returns hierarchy information points for one cohort and
// its direct child clusterQueues.
// Returns nil if the cohort does not exist or has a cycle.
func (m *Manager[CQ, C]) CohortInfoPoints(cohortName kueue.CohortReference) []InfoPoint {
	ch, found := m.cohortByName(cohortName)
	if !found || hasCycleIfSupported(ch) {
		return nil
	}
	rootCohortName := rootName(ch)
	return pointsForCohort(ch, rootCohortName)
}

// CohortSubtreeInfoPoints returns hierarchy information points for the subtree
// rooted at cohortName. For each cohort in the subtree, points include one
// cohort entry and one entry per direct child clusterQueue.
// Returns nil if the cohort does not exist or has a cycle.
func (m *Manager[CQ, C]) CohortSubtreeInfoPoints(cohortName kueue.CohortReference) []InfoPoint {
	subtreeRoot, found := m.cohortByName(cohortName)
	if !found || hasCycleIfSupported(subtreeRoot) {
		return nil
	}

	rootCohortName := rootName(subtreeRoot)
	stack := []C{subtreeRoot}
	var points []InfoPoint

	for len(stack) > 0 {
		last := len(stack) - 1
		ch := stack[last]
		stack = stack[:last]

		points = append(points, pointsForCohort(ch, rootCohortName)...)
		stack = append(stack, ch.ChildCohorts()...)
	}

	return points
}

func (m *Manager[CQ, C]) cohortByName(name kueue.CohortReference) (C, bool) {
	cohort := m.Cohort(name)
	var zero C
	return cohort, cohort != zero
}

func pointsForCohort[CQ clusterQueueNode[C], C cohortNode[CQ, C]](ch C, rootCohortName kueue.CohortReference) []InfoPoint {
	cohortName := ch.GetName()
	point := InfoPoint{
		CohortName:     &cohortName,
		RootCohortName: rootCohortName,
	}
	if ch.HasParent() {
		point.ParentCohortName = ch.Parent().GetName()
	}

	points := []InfoPoint{point}
	for _, cq := range ch.ChildCQs() {
		cqName := cq.GetName()
		points = append(points, InfoPoint{
			ClusterQueueName: &cqName,
			ParentCohortName: cohortName,
			RootCohortName:   rootCohortName,
		})
	}

	return points
}

func rootName[CQ clusterQueueNode[C], C cohortNode[CQ, C]](ch C) kueue.CohortReference {
	root := ch
	for root.HasParent() {
		root = root.Parent()
	}
	return root.GetName()
}

func hasCycleIfSupported[C any](cohort C) bool {
	cycleCheckable, ok := any(cohort).(CycleCheckable)
	return ok && HasCycle(cycleCheckable)
}
