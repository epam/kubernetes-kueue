# pkg/util/wait/

Wait and retry utilities for polling and backoff.

## Key Functions

- `PollUntilContextTimeout(ctx, interval, timeout, immediate, condition)` — poll until condition is true or timeout
- `ExponentialBackoff(opts, fn)` — retry with exponential backoff
- `RetryOnConflict(fn)` — retry Kubernetes API calls on conflict error

## Usage

```go
// Wait for a workload to be admitted:
err := utilwait.PollUntilContextTimeout(ctx, 100*time.Millisecond, 30*time.Second, true,
    func(ctx context.Context) (bool, error) {
        wl := &kueue.Workload{}
        if err := c.Get(ctx, key, wl); err != nil {
            return false, err
        }
        return workload.IsAdmitted(wl), nil
    })
```
