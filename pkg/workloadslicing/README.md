# pkg/workloadslicing/

Workload slicing support for elastic jobs. A `WorkloadSlice` represents an incremental resource request from an already-admitted workload that wants to scale up.

## Concept

```
Admitted Workload (8 GPUs)
  └── WorkloadSlice (+ 4 GPUs requested)
        → Kueue checks if 4 more GPUs are available
        → If yes: approves the slice, job scales to 12 GPUs
        → If no:  slice waits in queue
```

## Key Functions

- `NewWorkloadSlice(wl, delta)` — create a slice for a delta resource request
- `IsWorkloadSlice(obj)` — check if an object is a WorkloadSlice
- `AdmitSlice(ctx, c, slice)` — admit a pending slice
- `EvictSlice(ctx, c, slice, reason)` — revoke a slice admission

## Feature Gate

`features.ElasticJobsViaWorkloadSlices` — must be enabled for slicing to work.

## Interaction with Cache

WorkloadSlices are tracked separately in the cache. The parent workload's admission remains stable; only the slice's resources are added/removed from quota.
