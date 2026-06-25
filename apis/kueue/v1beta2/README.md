# apis/kueue/v1beta2/

Beta API version used internally by the scheduler and for some newer fields. The scheduler imports `v1beta2` types; controllers use `v1beta1` (hub version).

## Types Defined

This package mirrors `v1beta1` with additions/changes that have not yet been promoted to v1beta1:

- **`WorkloadPriorityClass`** — originates here before merging into v1beta1
- All v1beta1 types are also represented here with potential field extensions

## Conversion

Types here convert to/from `v1beta1` via `*_conversion.go` files. The Kubernetes API machinery automatically handles version negotiation during webhook processing and API serving.

## When to Use v1beta2

- Scheduler code (`pkg/scheduler/`) imports `kueue "sigs.k8s.io/kueue/apis/kueue/v1beta2"` for its internal representation
- New features may be added here first before being promoted to v1beta1

## Stability

`v1beta2` is Beta-stable. Changes may be made but will be backwards-compatible within the beta period.
