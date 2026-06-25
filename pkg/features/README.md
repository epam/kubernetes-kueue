# pkg/features/

Feature gate registry for Kueue. All feature gates are defined here using the `k8s.io/component-base/featuregate` framework.

## Key File: `kube_features.go`

Defines all feature gates with their default state and maturity level:

```go
var (
    TopologyAwareScheduling = featuregate.Feature("TopologyAwareScheduling")
    MultiKueue              = featuregate.Feature("MultiKueue")
    ConcurrentAdmission     = featuregate.Feature("ConcurrentAdmission")
    // ... 50+ more
)
```

## Feature Gate Lifecycle

| Stage | Default | Can Disable? | Notes |
|---|---|---|---|
| Alpha | Off | Yes | Opt-in, unstable API |
| Beta | On | Yes | Stable API, on by default |
| GA | On | No | Permanently enabled |

## Checking Feature Gates

```go
if features.Enabled(features.TopologyAwareScheduling) {
    // TAS-specific code path
}
```

## Setting Feature Gates

In the `Configuration`:
```yaml
featureGates:
  TopologyAwareScheduling: true
  ConcurrentAdmission: false
```

Or via controller manager flags:
```
--feature-gates=TopologyAwareScheduling=true,MultiKueue=true
```

## Notable Feature Gates

| Gate | Stage | Description |
|---|---|---|
| `TopologyAwareScheduling` | Beta | TAS placement constraints |
| `MultiKueue` | Beta | Multi-cluster job dispatch |
| `ConcurrentAdmission` | Alpha | Parallel flavor pursuit |
| `PartialAdmission` | Beta | Admit with fewer replicas |
| `LocalQueueDefaulting` | Beta | Auto-assign default queue |
| `ElasticJobsViaWorkloadSlices` | Alpha | Elastic job support |
| `FailureRecovery` | Alpha | Auto-retry on infra failure |
| `WorkloadDispatcher` | Alpha | Auto-route workloads to queues |
| `AdmissionFairSharing` | Alpha | Fair queue ordering |
