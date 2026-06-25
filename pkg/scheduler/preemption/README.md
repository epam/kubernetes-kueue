# pkg/scheduler/preemption/

Orchestrates workload preemption: determining which currently-admitted workloads must be evicted to make room for a higher-priority or fair-share-deserving incoming workload.

## Overview

Preemption runs when flavor assignment returns `Preempt` mode. The preemption package:
1. Identifies candidate victims (workloads that could be evicted)
2. Selects the minimal set of victims to free enough quota
3. Issues eviction by removing `status.admission` from victim Workloads
4. Waits for victims to terminate (via expectations) before admitting the preemptor

## Sub-packages

| Package | Purpose |
|---|---|
| [`classical/`](classical/) | Priority-based preemption (higher priority evicts lower) |
| [`fairsharing/`](fairsharing/) | DRS-based preemption (balance resource usage across CQs) |
| [`common/`](common/) | Shared utilities (candidate selection, victim ordering) |
| [`expectations/`](expectations/) | Tracks expected eviction completions to avoid re-preempting |

## Preemption Strategies

Configured via `ClusterQueue.spec.preemption`:

```yaml
preemption:
  reclaimWithinCohort: Any       # reclaim borrowed quota from anyone
  borrowWithinCohort:
    policy: LowerPriority        # borrow from lower-priority workloads
  withinClusterQueue: LowerPriority
```

## Victim Selection Algorithm

1. Collect all candidates: workloads in the same CQ (within-CQ preemption) and in cohort (cross-CQ borrowing)
2. Filter by priority rules (only evict lower or equal priority, depending on config)
3. Sort by "least damage": prefer evicting workloads that free the most resources
4. Select minimum set sufficient to accommodate the preemptor

## Events

Every evicted workload gets a `Preempted` event with:
- Preemptor workload UID and job UID
- Preemption reason
- Preemptor/preemptee ClusterQueue paths
- Effective priority values
