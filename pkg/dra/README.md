# pkg/dra/

Dynamic Resource Allocation (DRA) integration. Bridges Kueue's resource tracking with Kubernetes DRA (`resource.k8s.io`) for claiming structured resources (GPUs, accelerators, FPGAs) via the DRA API.

## DRA vs. Classic Resources

Classic resources (`requests.nvidia.com/gpu: 1`) are opaque counts. DRA allows:
- Structured resource claims with specific attributes
- Shared resources between pods
- Fine-grained resource topology (NUMA, PCIe lanes)

## Kueue Integration

When a `Workload` requests DRA resources:
1. Kueue checks `ResourceClaim` availability in the scheduler
2. On admission, Kueue binds the `ResourceClaim` to the workload's pods
3. On eviction, Kueue releases the claim

## Key Functions

- `ExtractResourcesFromClaims(wl, claims)` — compute effective resource usage from DRA claims
- `SyncClaims(ctx, c, wl)` — ensure ResourceClaims are consistent with workload state

## Feature Gate

`features.DynamicResourceAllocation` — when disabled, DRA fields are ignored.
