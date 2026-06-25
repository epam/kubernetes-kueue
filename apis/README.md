# apis/

This directory contains all Kueue API type definitions organized by API group and version. These are the Go structs that get serialized into Kubernetes CRDs and configuration objects.

## API Groups

| Group | Path | Purpose |
|---|---|---|
| `kueue.x-k8s.io` | [`kueue/`](kueue/) | Core Kueue CRDs (ClusterQueue, LocalQueue, Workload, etc.) |
| `config.kueue.x-k8s.io` | [`config/`](config/) | Controller manager configuration |
| `visibility.kueue.x-k8s.io` | [`visibility/`](visibility/) | Virtual API for inspecting queue state |

## Structure

```
apis/
├── config/
│   ├── v1beta1/   # Deprecated controller config (still supported)
│   └── v1beta2/   # Current controller config
├── kueue/
│   ├── v1alpha1/  # Alpha CRDs (Topology)
│   ├── v1beta1/   # Primary hub version for all CRDs
│   └── v1beta2/   # Newer CRDs (used as internal version by scheduler)
└── visibility/
    ├── v1beta1/   # PendingWorkloads virtual resource (v1)
    └── v1beta2/   # PendingWorkloads + LocalQueuePendingWorkloads (v2)
```

## Versioning and Conversion

Kueue uses `v1beta1` as the storage (hub) version for most CRDs. `v1beta2` types are converted to/from `v1beta1` via conversion functions (`*_conversion.go`). The scheduler internally uses `v1beta2` types.

Generated files (`zz_generated.*.go`) are produced by `controller-gen` and should never be edited manually.

## Key Design Principles

- All public API fields have kubebuilder validation markers (`+kubebuilder:validation:...`)
- CRD fields use `omitempty` JSON tags for forward/backward compatibility
- Status fields are updated via `status` subresource (separate PATCH path from spec)
- `DeepCopyObject` methods are auto-generated and required for controller-runtime
