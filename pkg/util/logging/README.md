# pkg/util/logging/

Structured logging helpers built on top of `k8s.io/klog/v2`.

## Verbosity Levels

Kueue follows these conventions:
| Level | Usage |
|---|---|
| `V(2)` | Major lifecycle events (workload admitted, evicted) |
| `V(3)` | Controller reconcile starts/ends |
| `V(4)` | Per-cycle details (flavor assignment result, queue state) |
| `V(5)+` | Trace-level debugging |

## Key Functions

- `KObj(obj)` — wraps `klog.KObj` for structured key-value pairs
- `WithFields(logger, keysAndValues...)` — create a logger with preset fields (e.g., workload name)

## Usage Convention

```go
log := ctrl.LoggerFrom(ctx).WithValues("workload", klog.KObj(wl))
log.V(2).Info("Workload admitted", "clusterQueue", cqName)
log.V(4).Info("Flavor assignment result", "mode", assignment.RepMode)
```

Do not use `log.Info()` (V(0)) for per-reconcile events — it pollutes logs in production.
