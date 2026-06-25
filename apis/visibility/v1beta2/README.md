# apis/visibility/v1beta2/

v1beta2 of the Visibility API. Adds `LocalQueuePendingWorkloads` for namespace-scoped queue inspection, and improves the `PendingWorkloads` resource.

## Types

### `PendingWorkloadsSummary`

Returned per-ClusterQueue. Same structure as v1beta1 with potential additional fields.

### `LocalQueuePendingWorkloadsSummary`

New in v1beta2. Scoped to a `LocalQueue` in a namespace.

```bash
# Get pending workloads in a local queue (namespace-scoped RBAC)
kubectl get localqueuependingworkloads my-queue -n my-namespace \
  --subresource=pendingworkloads
```

Fields:
```go
type LocalQueuePendingWorkload struct {
    metav1.ObjectMeta
    Priority              int32
    PositionInLocalQueue  int32
}
```

## RBAC

`LocalQueuePendingWorkloads` can be accessed with just namespace-scoped permissions, making it suitable for non-admin users to inspect their own queue positions.

## defaults.go

Sets default values for `PendingWorkloadOptions` (e.g., default limit on returned items).
