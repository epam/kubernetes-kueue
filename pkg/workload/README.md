# pkg/workload/

Workload abstraction layer. Provides `workload.Info` — a rich wrapper around the Kueue `Workload` API object with pre-computed fields for use by the scheduler and controllers.

## Key Type: `Info`

```go
type Info struct {
    Obj             *kueue.Workload       // the API object
    TotalRequests   []PodSetResources     // pre-computed total resources per PodSet
    CanBePartiallyAdmitted bool           // supports partial admission
    // ... additional computed fields
}
```

`Info` is created once per workload and passed through the scheduling pipeline. Controllers use `Obj` to update Kubernetes; the scheduler uses `TotalRequests` to avoid re-computing resources every cycle.

## Key Functions

### `workload.NewInfo(wl)`

Creates an `Info` from a `Workload` API object. Computes `TotalRequests` by iterating PodSets and their resource requests.

### `workload.EnsureQueue(wl, lq)`

Validates the workload's queue reference and ensures the LocalQueue exists.

### `workload.HasAdmission(wl)` / `workload.IsAdmitted(wl)`

Quick checks on workload admission state.

### `workload.EvictWorkload(ctx, c, wl, reason, msg)`

Removes `status.admission` and sets the `Evicted` condition. Used by the preemption and workload controllers.

## Sub-packages

| Package | Purpose |
|---|---|
| [`concurrentadmission/`](concurrentadmission/) | Helpers for concurrent admission (parent/child workload tracking) |

## Admission Status

```go
type Admission struct {
    ClusterQueue          ClusterQueueReference
    PodSetAssignments     []PodSetAssignment  // flavor + topology per PodSet
}
```

Written by the scheduler; read by job adapters to inject node affinity.
