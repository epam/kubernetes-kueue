# pkg/util/parallelize/

Parallel execution helpers for running operations concurrently with error collection.

## Key Functions

- `Until(ctx, parallelism int, work []T, fn func(T) error) []error` — run `fn` on each item in parallel, collect all errors
- `UntilFirstError(ctx, parallelism int, work []T, fn func(T) error) error` — stop on first error

## Usage

Used in the MultiKueue controller to fan out workload sync operations to multiple worker clusters concurrently:

```go
errs := parallelize.Until(ctx, 4, clusters, func(c *ClusterClient) error {
    return c.SyncWorkload(ctx, wl)
})
```
