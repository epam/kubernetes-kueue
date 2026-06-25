# apis/visibility/v1beta1/

v1beta1 of the Visibility API. Provides `PendingWorkloads` — a virtual list resource for inspecting the pending workload queue of a `ClusterQueue`.

## Types

### `PendingWorkloadsSummary`

Returned by `GET /apis/visibility.kueue.x-k8s.io/v1beta1/clusterqueues/{name}/pendingworkloads`.

```go
type PendingWorkloadsSummary struct {
    Items []PendingWorkload
}

type PendingWorkload struct {
    metav1.ObjectMeta  // Name = workload name, Namespace = workload namespace
    Priority                int32
    PositionInClusterQueue  int32
    PositionInLocalQueue    int32
    LocalQueueName          string
}
```

## Limitations

- Only returns up to `queueVisibility.clusterQueues.maxCount` workloads (configured in `Configuration`)
- Requires `get` permission on `clusterqueues/pendingworkloads`

## Deprecation

Use `v1beta2` for new integrations — it adds `LocalQueuePendingWorkloads` and better namespace scoping.
