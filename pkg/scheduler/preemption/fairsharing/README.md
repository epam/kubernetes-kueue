# pkg/scheduler/preemption/fairsharing/

Fair Sharing preemption strategy. Instead of priority alone, eviction decisions are based on DRS (Dominant Resource Share) — ClusterQueues using more than their fair share can have workloads evicted.

## DRS Formula

```
DRS(CQ) = max over resources {
    (usage[r] + request[r] - nominalQuota[r]) / lendable[r]
} / weight
```

Where:
- `usage[r]` = currently used resources of type `r` in the CQ
- `request[r]` = resources the incoming workload needs
- `nominalQuota[r]` = the CQ's guaranteed quota
- `lendable[r]` = total resources available for borrowing across the cohort
- `weight` = the CQ's fair sharing weight (default 1)

A higher DRS means the CQ is using more than its fair share.

## Preemption Decision

The incoming workload's CQ (preemptor) can preempt workloads from another CQ (victim CQ) if:
1. The victim CQ's DRS > preemptor CQ's DRS (victim is hoarding more resources)
2. Evicting a workload from the victim CQ would bring sufficient resources to admit the preemptor

## Configuration

```yaml
# In kueue Configuration:
fairSharing:
  enable: true
  preemptionStrategies:
  - LessThanOrEqualToFinalShare  # only preempt if preemptor ends up at or below victim's share
  - LessThanInitialShare         # only preempt if preemptor starts with less than victim
```
