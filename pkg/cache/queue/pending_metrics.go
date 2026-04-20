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

package queue

import (
	"context"
	"time"

	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"

	kueue "sigs.k8s.io/kueue/apis/kueue/v1beta2"
)

const pendingWaitMetricsPeriodicRefresh = 15 * time.Second

// PendingWaitMetricsNotifier is notified when pending wait gauge metrics for a ClusterQueue should be recomputed.
type PendingWaitMetricsNotifier interface {
	NotifyClusterQueue(kueue.ClusterQueueReference)
}

// NoopPendingWaitMetricsNotifier disables asynchronous pending-wait gauge refresh (for unit tests).
type NoopPendingWaitMetricsNotifier struct{}

func (NoopPendingWaitMetricsNotifier) NotifyClusterQueue(kueue.ClusterQueueReference) {}

// PendingMetricsController recomputes expensive pending wait-time gauges off the scheduling hot path.
// It also periodically refreshes all ClusterQueues so wait times track clock advancement without queue events.
type PendingMetricsController struct {
	manager       *Manager
	queue         workqueue.TypedDelayingInterface[kueue.ClusterQueueReference]
	batchPeriod   time.Duration
	refreshPeriod time.Duration
}

func newPendingMetricsController(m *Manager) *PendingMetricsController {
	return &PendingMetricsController{
		manager:       m,
		queue:         workqueue.NewTypedDelayingQueue[kueue.ClusterQueueReference](),
		batchPeriod:   getRequeueBatchPeriod(),
		refreshPeriod: pendingWaitMetricsPeriodicRefresh,
	}
}

// NewPendingMetricsController returns the Runnable that updates pending wait gauge metrics asynchronously.
// It returns nil when the Manager was built with a noop pending-wait notifier (unit tests) or m is nil.
func NewPendingMetricsController(m *Manager) *PendingMetricsController {
	if m == nil {
		return nil
	}
	return m.pendingMetricsController
}

// NotifyClusterQueue schedules a delayed refresh of pending wait gauges for the ClusterQueue.
func (r *PendingMetricsController) NotifyClusterQueue(cqName kueue.ClusterQueueReference) {
	r.queue.AddAfter(cqName, r.batchPeriod)
}

// Start runs the workqueue consumer and a periodic full refresh until ctx is cancelled.
func (r *PendingMetricsController) Start(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("pending_wait_metrics_worker")
	ctx = ctrl.LoggerInto(ctx, log)
	go func() {
		<-ctx.Done()
		r.queue.ShutDown()
	}()
	go func() {
		ticker := time.NewTicker(r.refreshPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.refreshAllWaitGaugeMetrics()
			}
		}
	}()
	for {
		item, shutdown := r.queue.Get()
		if shutdown {
			return nil
		}
		r.reconcileOne(item)
		r.queue.Done(item)
	}
}

func (r *PendingMetricsController) reconcileOne(cqName kueue.ClusterQueueReference) {
	r.manager.Lock()
	defer r.manager.Unlock()
	cq := r.manager.hm.ClusterQueue(cqName)
	if cq == nil {
		return
	}
	reportCQPendingWaitGaugeMetrics(r.manager, cq)
}

func (r *PendingMetricsController) refreshAllWaitGaugeMetrics() {
	r.manager.Lock()
	defer r.manager.Unlock()
	for _, cq := range r.manager.hm.ClusterQueues() {
		reportCQPendingWaitGaugeMetrics(r.manager, cq)
	}
}
