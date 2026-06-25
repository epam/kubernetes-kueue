# pkg/visibility/storage/

REST storage backends for the Kueue Visibility API. Implements the `k8s.io/apiserver/pkg/registry/rest` interfaces to serve pending workload data from the in-memory queue cache.

## Key Types

### `pendingWorkloadsInCQREST`

Serves `GET clusterqueues/{name}/pendingworkloads` by querying `pkg/cache/queue.Manager` and returning a sorted list of `PendingWorkload` objects.

### `pendingWorkloadsInLQREST`

Serves `GET namespaces/{ns}/localqueues/{name}/pendingworkloads` with namespace-scoped results.

## Sorting

Results are sorted by:
1. Priority (descending — highest priority first)
2. Creation timestamp (ascending — oldest first within same priority)

## Depth Limit

Returns at most `queueVisibility.clusterQueues.maxCount` workloads (from `Configuration`). Default: 10.
