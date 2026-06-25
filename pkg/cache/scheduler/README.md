# pkg/cache/scheduler/

Scheduler-facing view of the Kueue cache. Wraps the core cache implementation and exposes the `Snapshot()` API used exclusively by the scheduler.

## Key Type: `Cache`

Thin wrapper around the core cache with:
- `Snapshot()` — produces an immutable `Snapshot` of all ClusterQueues, Cohorts, and ResourceFlavors for one scheduler cycle
- All mutation methods (inherited) — called by controllers to update cache state

## Snapshot Structure

```go
type Snapshot struct {
    ClusterQueues        map[string]*ClusterQueue
    ResourceFlavors      map[kueuev1b1.ResourceFlavorReference]*kueuev1b1.ResourceFlavor
    InactiveClusterQueues sets.Set[string]
    // ...
}
```

The snapshot is a deep copy — the scheduler is free to mutate it (e.g., to simulate admitting a workload) without affecting the live cache. This is key to the lock-free scheduler design.

## Usage Pattern

```
Scheduler.schedule() {
    snap := cache.Snapshot()         // single lock-protected copy
    heads := queues.Heads(ctx)       // pending workloads
    for _, head := range heads {
        flavorassigner.Assign(snap, head)  // no locks needed
        preemption.Run(snap, head)
    }
}
```
