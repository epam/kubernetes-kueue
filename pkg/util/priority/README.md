# pkg/util/priority/

Workload priority helpers. Computes effective priority for a workload from its `WorkloadPriorityClass` and Kubernetes `PriorityClass`.

## Key Functions

- `Priority(wl *kueue.Workload) int32` — get the effective integer priority
- `PriorityClassSource(wl) PriorityClassSource` — which source set the priority (WorkloadPriorityClass vs. PriorityClass)
- `IssueOrdering(a, b *workload.Info) bool` — compare two workloads for queue ordering (higher priority first, older first within same priority)

## Priority Sources

1. `WorkloadPriorityClass` — Kueue-specific priority, preferred
2. Kubernetes `PriorityClass` — fallback via pod template priority class name
3. Default priority: 0

## Usage

The queue manager and preemption logic use `priority.Priority(wl)` to determine ordering and eviction targets.
