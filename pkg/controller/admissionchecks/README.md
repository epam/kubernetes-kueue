# pkg/controller/admissionchecks/

Framework for pluggable external admission gates. An `AdmissionCheck` allows a third-party controller to approve or deny workload admission before Kueue allows a job to run.

## Concept

A `ClusterQueue` can reference one or more `AdmissionCheck` objects. When a workload is admitted to quota, it enters a "reservation" phase and waits for all referenced `AdmissionCheck`s to set their status to `Ready`. Only then does Kueue unsuspend the underlying job.

```
Workload admitted → QuotaReserved
  → AdmissionCheck A: Pending → Ready ✓
  → AdmissionCheck B: Pending → Ready ✓
  → Job unsuspended (runs)
```

## Built-in Implementations

| Controller | Package | Purpose |
|---|---|---|
| `MultiKueue` | [`multikueue/`](multikueue/) | Dispatch workload to a worker cluster |
| `ProvisioningRequest` | [`provisioning/`](provisioning/) | Request node provisioning from Cluster Autoscaler |

## Custom Admission Checks

External controllers can implement their own `AdmissionCheck` by:
1. Creating an `AdmissionCheck` CRD object with `spec.controllerName = "my-controller"`
2. Watching `Workload` objects that reference the CQ
3. Setting `workload.status.admissionChecks[].state = Ready` when satisfied

## Key Types (in `apis/kueue/v1beta1/admissioncheck_types.go`)

```go
type AdmissionCheckSpec struct {
    ControllerName string               // which controller owns this check
    RetryDelayMinutes *int64            // min delay between retries
    Parameters *AdmissionCheckParametersReference // per-check config
}

// On a Workload:
type AdmissionCheckState struct {
    Name    string
    State   CheckState  // Pending / Ready / Retry
    Message string
}
```
