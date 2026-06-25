# pkg/controller/

All Kueue controller-runtime reconcilers. Each subdirectory is a self-contained controller or set of controllers for a specific resource type or feature.

## Structure

```
controller/
├── core/                  # ClusterQueue, LocalQueue, Workload, ResourceFlavor reconcilers
├── admissionchecks/       # AdmissionCheck framework controllers
│   ├── multikueue/        # MultiKueue admission check (multi-cluster dispatch)
│   └── provisioning/      # ProvisioningRequest admission check (Cluster Autoscaler)
├── concurrentadmission/   # ConcurrentAdmission — parallel flavor pursuit
├── elasticjobs/           # Elastic job / WorkloadSlice support
├── failurerecovery/       # Failure recovery for admitted workloads
├── jobframework/          # GenericJob plugin interface and base reconciler
├── jobs/                  # Per-framework job adapters (batch/Job, JobSet, Ray, etc.)
├── tas/                   # Topology-Aware Scheduling controller
└── workloaddispatcher/    # WorkloadDispatcher for routing workloads
```

## Common Patterns

All controllers:
- Embed `controller-runtime`'s `Reconciler` interface
- Use structured logging via `logr` / `klog`
- Write status via `status` subresource to avoid conflicts
- Use `Owns()` / `Watches()` to trigger on relevant resource changes
- Record Kubernetes events for significant state transitions

## Controller Registration

Controllers are registered in `cmd/kueue/main.go` via `Setup()` functions in each package. The `jobframework.IntegrationManager` dynamically registers job-type controllers based on enabled integrations in the `Configuration`.
