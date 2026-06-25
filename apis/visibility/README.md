# apis/visibility/

Virtual API types for the `visibility.kueue.x-k8s.io` group. These are **not CRDs stored in etcd** — they are served by a custom API server registered as an aggregated API extension (`APIService`).

## Purpose

Allows users to inspect live queue state via standard `kubectl get` commands:

```bash
kubectl get pendingworkloads -n team-a
kubectl get localqueuependingworkloads my-queue -n team-a
```

## Versions

| Version | Resources |
|---|---|
| `v1beta1` | `PendingWorkloads` (cluster-queue scoped) |
| `v1beta2` | `PendingWorkloads` + `LocalQueuePendingWorkloads` |

## Key Types

### `PendingWorkload` (v1beta1/v1beta2)

Represents a single workload waiting in a queue. Fields:
- `priority` — effective priority
- `positionInClusterQueue` / `positionInLocalQueue` — queue position
- `localQueueName` — which local queue submitted it

### `LocalQueuePendingWorkloads` (v1beta2)

A list resource scoped to a `LocalQueue`. Returns pending workloads visible only within that namespace/queue combination, respecting RBAC.

## Implementation

The API server is implemented in `pkg/visibility/` and registered via `APIService` in the Kubernetes aggregation layer. It does NOT use etcd storage — queries go directly to the in-memory `pkg/cache/queue` manager.

## openapi/

Contains the OpenAPI schema for the visibility API, used for generating client code and documentation.
