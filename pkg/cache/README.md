# pkg/cache/

The central in-memory state store for the Kueue scheduler. Tracks all `ClusterQueue`s, `Cohort`s, admitted `Workload`s, and current resource usage. Provides the scheduler with a consistent point-in-time snapshot for making admission decisions without holding locks during scheduling.

## Responsibilities

- Maintain an accurate in-memory model of every `ClusterQueue` (quota, usage, admitted workloads)
- Build the hierarchical cohort tree for borrowing/lending quota
- Serve point-in-time snapshots to the scheduler loop
- Propagate queue/workload state changes from controllers

## Key Types

### `Cache` (pkg/cache/scheduler/cache.go)

The primary cache exposed to the scheduler. Methods:
- `Snapshot()` — returns a point-in-time `Snapshot` of all ClusterQueues and Cohorts
- `AddOrUpdateClusterQueue(cq)` / `DeleteClusterQueue(key)`
- `AddOrUpdateWorkload(wl)` / `ForgetWorkload(wl)`
- `AddOrUpdateLocalQueue(lq)` / `DeleteLocalQueue(key)`
- `AddOrUpdateResourceFlavor(rf)` / `DeleteResourceFlavor(key)`

### `Snapshot`

An immutable copy of the cache state at a single point in time. Used by the scheduler loop so it can iterate queues and test admission decisions without locking the live cache. Contains:
- `ClusterQueues` — map of ClusterQueue snapshots
- `ResourceFlavors` — available flavors
- `InactiveClusterQueueSets` — CQs that cannot admit

### ClusterQueue cache model

Each `ClusterQueue` tracks:
- `allocatableResourceGeneration` — monotonic counter for snapshot invalidation
- `usage` — current reserved resources per flavor/resource
- `workloads` — all admitted workloads
- `localQueues` — associated LocalQueues with their workload sets

## Sub-packages

| Package | Purpose |
|---|---|
| [`hierarchy/`](hierarchy/) | Cohort tree (parent-child relationships, borrowable quota aggregation) |
| [`queue/`](queue/) | Pending workload queue ordering (FIFO, FairSharing) |
| [`queue/afs/`](queue/afs/) | Admission Fair Sharing queue ordering algorithm |
| [`scheduler/`](scheduler/) | Scheduler-facing cache view and snapshot |

## Thread Safety

The cache uses a `sync.RWMutex`. The scheduler holds a read lock only while copying the snapshot; writes happen in controller reconcilers. This means admission decisions are made on a snapshot with no lock contention.
