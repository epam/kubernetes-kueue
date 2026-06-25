# pkg/metrics/

Prometheus metrics definitions for Kueue. All metrics are registered here and recorded by controllers and the scheduler.

## Metric Categories

### ClusterQueue Metrics

Counters/gauges scoped to a specific `ClusterQueue`:

| Metric | Type | Description |
|---|---|---|
| `kueue_admitted_workloads_total` | Counter | Workloads admitted (by CQ, reason) |
| `kueue_evicted_workloads_total` | Counter | Workloads evicted (by CQ, reason) |
| `kueue_pending_workloads` | Gauge | Currently pending workloads |
| `kueue_admitted_active_workloads` | Gauge | Currently admitted workloads |
| `kueue_cluster_queue_resource_usage` | Gauge | Used resources per flavor |
| `kueue_cluster_queue_nominal_quota` | Gauge | Configured nominal quota |
| `kueue_cluster_queue_borrowing_limit` | Gauge | Configured borrowing limit |
| `kueue_cluster_queue_lending_limit` | Gauge | Configured lending limit |
| `kueue_cluster_queue_status` | Gauge | CQ status (active=1, inactive=0) |

### Scheduler Metrics

| Metric | Type | Description |
|---|---|---|
| `kueue_admission_cycle_preemption_skips` | Counter | Cycles skipped due to preemption |
| `kueue_admission_attempts_total` | Counter | Total admission attempts |
| `kueue_admission_wait_time_seconds` | Histogram | Time from submit to admit |
| `kueue_admission_checks_wait_time_seconds` | Histogram | Time waiting for admission checks |

### Workload Metrics

| Metric | Type | Description |
|---|---|---|
| `kueue_workload_evictions_total` | Counter | Total evictions |
| `kueue_workload_create_time_seconds` | Histogram | Workload creation latency |

### Local Queue Metrics

Per-`LocalQueue` mirrors of CQ metrics (admitted, pending, resource usage).

### TAS Metrics

TAS-specific metrics for topology assignment decisions.

## Custom Labels

The `custom_labels.go` file implements the `CustomLabels` feature gate — allows operators to add configurable extra labels to all metrics (e.g., team, environment).

## Registration

All metrics are registered via `prometheus.MustRegister()` at init time. The metrics endpoint is served on the controller manager's metrics port (default: `:8080`).
