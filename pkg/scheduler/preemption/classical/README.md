# pkg/scheduler/preemption/classical/

Classical (priority-based) preemption strategy. A workload with higher priority can evict workloads with lower priority to claim resources.

## Algorithm

1. Identify all admitted workloads in the same `ClusterQueue` (and optionally cohort members) as victims
2. Filter to only those with lower effective priority than the preemptor
3. Sort victims: prefer evicting the lowest-priority, most-recently-admitted workloads first (minimizes disruption)
4. Greedily select victims until enough resources are freed

## Configuration

```yaml
spec:
  preemption:
    withinClusterQueue: LowerPriority     # evict lower-priority in same CQ
    reclaimWithinCohort: Any              # reclaim borrowed quota from any CQ in cohort
    borrowWithinCohort:
      policy: LowerPriority              # borrow only by evicting lower-priority
      maxPriorityThreshold: 100          # only borrow from workloads below this priority
```

## Relation to FairSharing Preemption

Classical preemption is triggered when `FairSharing.Enable = false` or when the `withinClusterQueue` rule applies. FairSharing preemption (`fairsharing/`) is used when the fair sharing strategy is `Preempt` and a workload's DRS exceeds the threshold.
