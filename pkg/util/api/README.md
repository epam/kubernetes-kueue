# pkg/util/api/

Kubernetes API helpers for common client operations.

## Key Functions

- `UpdateStatus(ctx, c, obj, mutate)` — update an object's status with automatic conflict retry
- `Patch(ctx, c, obj, mutate)` — patch an object with conflict retry
- `IgnoreNotFound(err) error` — suppress `IsNotFound` errors (common in controllers)
- `EqualSlices(a, b)` — compare two slices for equality in API contexts

## Usage

```go
// Update workload status with retry on conflict:
if err := utilapi.UpdateStatus(ctx, r.client, &wl, func() {
    wl.Status.Conditions = append(wl.Status.Conditions, admittedCondition)
}); err != nil {
    return err
}
```
