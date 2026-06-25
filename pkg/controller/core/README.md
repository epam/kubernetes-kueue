# pkg/controller/core/

Core Kueue controllers that reconcile the fundamental CRDs: `ClusterQueue`, `LocalQueue`, `Workload`, `ResourceFlavor`, `Cohort`, and `WorkloadPriorityClass`.

## Controllers

### `ClusterQueueReconciler` (`clusterqueue_controller.go`)

Watches `ClusterQueue` objects and:
- Validates referenced `ResourceFlavor`s exist and are active
- Validates referenced `AdmissionCheck`s exist and are active
- Updates `ClusterQueue.status.conditions[Active]` based on validity
- Propagates queue stop/drain state
- Triggers scheduler re-evaluation when quotas change

### `LocalQueueReconciler` (`localqueue_controller.go`)

Watches `LocalQueue` objects and:
- Links LocalQueue to its ClusterQueue in the cache
- Propagates ClusterQueue status (flavor usage, pending counts) to LocalQueue status
- Handles `stopPolicy` — holds or drains pending workloads

### `WorkloadReconciler` (`workload_controller.go`)

The most complex controller. Watches `Workload` objects and:
- Dequeues workloads that are admitted (schedules them in the cache)
- Removes `status.admission` when a workload finishes or is evicted
- Manages the `Finished`, `Evicted`, `PodsReady`, `QuotaReserved` conditions
- Triggers preemption expectations cleanup
- Handles `maximumExecutionTime` deadline enforcement

### `ResourceFlavorReconciler` (`resourceflavor_controller.go`)

- Validates `ResourceFlavor` node labels and taints exist on actual cluster nodes
- Updates `ResourceFlavor.status.conditions[Active]`

### `CohortReconciler` (`cohort_controller.go`)

- Validates `Cohort` objects
- Maintains the cohort hierarchy in the cache

### `WorkloadPriorityClassReconciler` (`workloadpriorityclass_controller.go`)

- Simple controller; validates and propagates `WorkloadPriorityClass` changes

### `AdmissionCheckController` (`admissioncheck_controller.go`)

- Watches `AdmissionCheck` objects and updates their active state

### `ResourceSliceController` (`resourceslice_controller.go`)

- DRA integration: watches `ResourceSlice` objects for Dynamic Resource Allocation

## Sub-packages

| Package | Purpose |
|---|---|
| [`indexer/`](indexer/) | Kubernetes informer field indexes for efficient cache lookups |

## Key Files

- `core.go` — register all core controllers with the manager
- `helpers.go` — shared reconciliation utilities
- `leader_aware_reconciler.go` — base for leader-election-aware controllers
