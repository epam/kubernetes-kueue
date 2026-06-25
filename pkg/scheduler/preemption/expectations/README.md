# pkg/scheduler/preemption/expectations/

Tracks pending preemption expectations to avoid re-preempting the same victims before they have finished terminating.

## Problem

After issuing an eviction (removing a workload's admission), the victim workload is still running — its pods take time to terminate. If the scheduler runs another cycle immediately, it might try to preempt the same workload again (it still shows as "admitted" in the cache until termination completes).

## Solution

The expectations package records which workloads have been issued eviction notices. The scheduler checks these expectations before selecting new victims:

```go
if expectations.SatisfiedFor(preemptor) {
    // All previously-evicted victims have terminated; safe to run
}
```

Expectations are cleared when:
1. The victim workload's `status.admission` is removed by the workload controller
2. The victim workload is deleted

## Implementation

Uses a TTL-based map: `preemptor UID → set of victim UIDs expected to terminate`. Each entry expires after a configurable timeout to avoid deadlocks if a victim fails to terminate.
