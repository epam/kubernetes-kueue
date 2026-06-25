# pkg/scheduler/

The core scheduling loop that decides which pending workloads to admit and with which resource flavors. Runs as a goroutine inside the controller manager, independent of controller reconcilers.

## Scheduling Cycle

The scheduler runs a continuous loop (`schedule()`) with these phases:

```
1. Heads        — ask queue.Manager for the head workload of each ClusterQueue
2. Snapshot     — take a point-in-time copy of cache state (lock-free after this)
3. Nominate     — for each head, check if it fits or can be preempted
4. FlavorAssign — run FlavorAssigner to find the best resource flavor assignment
5. Preempt      — if needed, evict lower-priority workloads to make room
6. Admit        — write Admission to the Workload's status in the API server
7. Requeue      — workloads that couldn't be admitted go back to the queue
```

## Key Types

### `Scheduler`

```go
type Scheduler struct {
    queues   *qcache.Manager        // pending workload queue
    cache    *schdcache.Cache       // in-memory cluster state
    client   client.Client          // API server client
    recorder events.EventRecorder   // Kubernetes events
    // ...
}
```

### `entry`

Per-workload state within a scheduling cycle:
```go
type entry struct {
    workload.Info
    assignment      flavorassigner.Assignment  // result of flavor assignment
    preemptionTargets []*workload.Info         // workloads to evict
}
```

## Fair Sharing Iterator

`fair_sharing_iterator.go` implements the FairSharing admission strategy: instead of admitting ClusterQueues in arrival order, it iterates them by DRS (Dominant Resource Share) so the CQ using the least resources gets the next admission attempt.

## Sub-packages

| Package | Purpose |
|---|---|
| [`flavorassigner/`](flavorassigner/) | Assign ResourceFlavors to workload PodSets |
| [`preemption/`](preemption/) | Decide which workloads to evict to make room |

## Concurrency

The scheduler loop is single-threaded by design — one workload is admitted per cycle. This simplifies correctness. The `ConcurrentAdmission` feature gate (`pkg/controller/concurrentadmission/`) runs multiple parallel flavor pursuit attempts for a single workload to reduce TOCTOU latency.

## Metrics

The scheduler records metrics for every cycle (admitted, queued, skipped workloads) via `pkg/metrics`.
