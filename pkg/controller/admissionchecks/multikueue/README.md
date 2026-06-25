# pkg/controller/admissionchecks/multikueue/

MultiKueue admission check controller. Implements the `AdmissionCheck` interface to dispatch workloads from a manager cluster to one of multiple worker clusters.

## Architecture

```
Manager Cluster                    Worker Cluster A
┌─────────────────────────┐       ┌──────────────────┐
│ Workload (admitted)     │──────▶│ Workload (copy)  │
│ AdmissionCheck: MK      │       │ Job (copy)       │
│   state: Pending        │       │ Pods (running)   │
└─────────────────────────┘       └──────────────────┘
          │
          ▼ (when worker job finishes)
   AdmissionCheck: Ready
   Workload: Finished
```

## Controllers

### `clustersReconciler`

Watches `MultiKueueCluster` objects. For each cluster:
- Reads the kubeconfig secret
- Creates/maintains a watch connection to the worker cluster
- Detects connectivity loss and marks the cluster inactive

### `wlReconciler`

Watches `Workload` objects that have a MultiKueue `AdmissionCheck`. For each:
- Selects a target worker cluster (based on cluster availability)
- Creates a copy of the Workload + parent Job on the worker
- Watches for the worker Workload to finish
- Syncs status back to the manager Workload
- Sets `AdmissionCheck.state = Ready` when ready / `Finished` when done

## Job Adapters (`MultiKueueAdapter`)

Each job type has a MultiKueue adapter implementing:
```go
type MultiKueueAdapter interface {
    SyncJob(ctx, localClient, remoteClient, key, workloadName, origin) error
    DeleteRemoteObjects(ctx, localClient, key) error
    IsJobManagedByKueue(ctx, c, key) (bool, error, string)
    KeepAdmissionCheckPending() bool
    GVK() schema.GroupVersionKind
}
```

## External Framework Support (`externalframeworks/`)

Custom job types can register MultiKueue adapters via `ExternalFrameworks` in the `Configuration`. These are discovered dynamically from CRD annotations.

## Key Labels

- `kueue.x-k8s.io/multikueue-origin` — marks objects created by the manager on worker clusters
- `kueue.x-k8s.io/multikueue-uid` — links worker objects back to manager workload
