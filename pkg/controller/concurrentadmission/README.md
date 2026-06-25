# pkg/controller/concurrentadmission/

ConcurrentAdmission controller. Enables parallel flavor pursuit — a single workload can simultaneously attempt admission to multiple `ResourceFlavor`s, reducing latency caused by sequential retries.

## Problem

In a large cluster with many flavors, a workload waiting for GPU resources might need to try many flavors sequentially. Each cycle takes time. If a workload has a preferred flavor that's temporarily unavailable, it waits multiple cycles.

## Solution

ConcurrentAdmission tracks "in-flight" admission attempts. The workload can be presented to the scheduler multiple times concurrently, each time with a different target flavor. The first successful admission wins.

## Key Types

### `Controller`

```go
type Controller struct {
    client client.Client
    cache  *schedulercache.Cache
}
```

Watches `Workload` objects and manages the in-flight set for each.

### `variants` function

Generates the list of flavor variants a workload can pursue in parallel. Takes the `ClusterQueue`'s `ResourceGroup`s and expands them into individual target assignments.

## Feature Gate

Enabled by `features.ConcurrentAdmission`. When disabled, the standard sequential scheduler behavior applies.

## Interaction with Scheduler

The scheduler checks `concurrentadmission.IsParent(wl)` before queueing workloads. Parent workloads (those with concurrent variants) are not re-queued by standard queue management — their lifecycle is managed by this controller.

## Helper Package

`pkg/workload/concurrentadmission/` contains workload-level helpers for managing concurrent admission state (parent/child workload relationships, retention logic).
