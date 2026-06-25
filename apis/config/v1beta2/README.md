# apis/config/v1beta2/

Current version of the Kueue controller manager configuration API.

## Key Types

### `Configuration`

The root configuration object. Loaded at startup from a file or ConfigMap.

**Important fields:**

| Field | Type | Description |
|---|---|---|
| `Namespace` | `string` | Namespace where Kueue runs |
| `ControllerManager` | embedded | Leader election, metrics, health probe config |
| `InternalCertManagement` | `*InternalCertManagement` | Auto-cert for webhooks (disable if using cert-manager) |
| `ClientConnection` | `*ClientConnection` | API server QPS/burst limits |
| `Integrations` | `*Integrations` | Which job frameworks to enable |
| `QueueVisibility` | `*QueueVisibility` | How many pending workloads to expose per queue |
| `MultiKueue` | `*MultiKueue` | Worker cluster GC interval, namespace |
| `FairSharing` | `*FairSharing` | Fair sharing strategy (`None`, `Preempt`) |
| `Resources` | `*Resources` | Resource transformer config, excluded prefixes |
| `AdmissionFairSharing` | `*AdmissionFairSharing` | Admission-time fair sharing config |

### `Integrations`

Controls which job frameworks Kueue manages:

```go
type Integrations struct {
    Frameworks         []string                   // e.g. "batch/v1", "ray.io/v1"
    ExternalFrameworks []ExternalFramework         // CRD-based plugins
    PodOptions         *PodIntegrationOptions      // Plain pod settings
    LabelKeysToCopy    []string                   // Labels copied from jobs to workloads
}
```

### `MultiKueue`

```go
type MultiKueue struct {
    GCInterval             *metav1.Duration  // How often to GC stale worker objects
    Origin                 *string           // Label value marking this as the manager cluster
    WorkerLostTimeout      *metav1.Duration  // When to declare a worker cluster lost
}
```

## Defaults

`defaults.go` applies default values when fields are omitted. Defaults are also registered via `zz_generated.defaults.go`.
