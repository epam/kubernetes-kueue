# pkg/workload/concurrentadmission/

Workload-level helpers for the ConcurrentAdmission feature. Manages the parent-child relationship between a "parent" workload and its flavor-pursuit "child" variants.

## Concept

When ConcurrentAdmission is enabled, a workload being considered for admission spawns multiple child workload variants (one per candidate flavor). These children race to be admitted; the first success wins and the rest are cleaned up.

## Key Functions

- `IsParent(wl)` — returns true if the workload is a parent (has active children)
- `IsChild(wl)` — returns true if the workload is a flavor-pursuit child
- `ParentName(wl)` — returns the parent workload name for a child
- `RetainFirstAdmission(ctx, c, parent, children)` — when one child is admitted, evict all others and promote the admission to the parent
- `CleanupChildren(ctx, c, parent)` — delete all child workloads for a parent

## Labels Used

- `kueue.x-k8s.io/concurrent-admission-parent` — set on child workloads to reference their parent
- `kueue.x-k8s.io/concurrent-admission-flavor` — the target flavor for this child attempt
