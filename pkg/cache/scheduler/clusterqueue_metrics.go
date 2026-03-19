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
	"github.com/go-logr/logr"

	kueue "sigs.k8s.io/kueue/apis/kueue/v1beta2"
	"sigs.k8s.io/kueue/pkg/metrics"
)

func (c *Cache) RecordClusterQueueResourceMetrics(log logr.Logger, cqName kueue.ClusterQueueReference, fairSharingEnabled bool) {
	if cqName == "" {
		log.V(4).Info("ClusterQueue name is empty, skipping resource metrics recording")
		return
	}

	log = c.withClusterQueueLogger(log, cqName)
	log.V(4).Info("Recording resource metrics for ClusterQueue")

	c.RLock()
	defer c.RUnlock()

	cq := c.hm.ClusterQueue(cqName)
	if cq == nil {
		return
	}

	cq.reportResourceMetrics(fairSharingEnabled)
}

func (c *Cache) ClearClusterQueueResourceMetrics(log logr.Logger, cqName kueue.ClusterQueueReference) {
	if cqName == "" {
		log.V(4).Info("ClusterQueue name is empty, skipping resource metrics clearing")
		return
	}

	log = c.withClusterQueueLogger(log, cqName)
	log.V(4).Info("Clearing resource metrics for ClusterQueue")

	metrics.ClearClusterQueueMetrics(cqName)
}

func (c *Cache) withClusterQueueLogger(log logr.Logger, cqName kueue.ClusterQueueReference) logr.Logger {
	return log.WithValues("clusterQueue", cqName)
}
