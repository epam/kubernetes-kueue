# pkg/util/waitforpodsready/

Wait-for-pods-ready logic. Implements the `PodsReady` condition tracking — Kueue can hold the next workload from running until the current workload's pods are all ready.

## Purpose

When `Configuration.waitForPodsReady.enable = true`, Kueue blocks admission of new workloads until the currently-admitted workload's pods enter `Ready` state. This prevents resource overcommit when pod startup is slow.

## Key Types

### `WaitForPodsReadyConfig`

```go
type WaitForPodsReadyConfig struct {
    Enable              bool
    Timeout             *metav1.Duration   // evict if pods not ready within timeout
    BlockAdmission      *bool              // block new admissions until current pods ready
    RequeuingStrategy   *RequeuingStrategy // how to requeue after timeout
}
```

### `RequeuingStrategy`

- `Backoff` — exponential backoff before requeuing
- `ImmediatelyUncapped` — requeue immediately, no cap on retries

## Key Functions

- `SatisfiedByPodsReady(wl, config) bool` — check if the pods-ready condition is satisfied given config
- `ShouldBlockAdmission(wl, queue) bool` — should the next workload wait?
