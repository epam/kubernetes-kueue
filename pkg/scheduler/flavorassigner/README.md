# pkg/scheduler/flavorassigner/

Assigns `ResourceFlavor`s to a workload's `PodSet`s during a scheduling cycle. This is the core resource-matching algorithm that determines whether a workload can be admitted to a ClusterQueue.

## Algorithm

For each `PodSet` in the workload, the `FlavorAssigner` iterates the `ResourceGroup`s of the target `ClusterQueue` in order and tries each `ResourceFlavor`:

```
for each ResourceGroup in ClusterQueue:
  for each ResourceFlavor in ResourceGroup:
    check: does this flavor have enough capacity?
    → Fit       (fits within nominal quota, no borrowing needed)
    → FitBorrowing (fits but requires borrowing from cohort)
    → Preempt   (fits after evicting lower-priority workloads)
    → NoFit     (cannot fit even with preemption)
```

### Assignment Modes

| Mode | Meaning |
|---|---|
| `Fit` | All resources fit within the CQ's nominal quota |
| `FitBorrowing` | Fits by borrowing unused quota from cohort members |
| `Preempt` | Fits after preempting other workloads |
| `NoFit` | Cannot be scheduled at this time |

## Key Types

### `FlavorAssigner`

```go
type FlavorAssigner struct {
    wl     *workload.Info
    cq     *cache.ClusterQueue
    snap   *cache.Snapshot
}
```

### `Assignment`

The result of flavor assignment:
```go
type Assignment struct {
    PodSets    []PodSetAssignment  // per-PodSet flavor + resource allocation
    Usage      resources.FlavorResourceQuantities
    RepMode    RepresentativeMode  // Fit / FitBorrowing / Preempt / NoFit
    // ...
}
```

### `PodSetAssignment`

```go
type PodSetAssignment struct {
    Name    string
    Flavors ResourceAssignment  // resource → flavor mapping
    Count   *int32
    TopologyAssignment *kueue.TopologyAssignment  // for TAS
}
```

## TAS Integration

When a `ResourceFlavor` references a `Topology`, the flavor assigner delegates to TAS topology assignment logic to ensure pods are placed on the appropriate topology nodes.

## Partial Admission

When `PartialAdmission` feature gate is enabled, the assigner can reduce `PodSet.count` below the requested count if a smaller batch fits and the workload opts in via `kueue.x-k8s.io/max-exec-time-seconds` annotation.
