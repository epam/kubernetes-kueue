# pkg/controller/jobs/kubeflow/kubeflowjob/

Shared base adapter for all KubeFlow Training Operator job types. Implements `GenericJob` for the common pattern across PyTorchJob, TFJob, JAXJob, PaddleJob, XGBoostJob.

## Key Type: `KubeflowJob`

```go
type KubeflowJob struct {
    object  KFJobObject  // the concrete KubeFlow job (PyTorchJob, TFJob, etc.)
}
```

### `KFJobObject` Interface

Each KubeFlow job type implements:
```go
type KFJobObject interface {
    client.Object
    RunPolicy() *kftraining.RunPolicy
    ReplicaSpecs() map[kftraining.ReplicaType]*kftraining.ReplicaSpec
}
```

## GenericJob Implementation

| Method | Behavior |
|---|---|
| `IsSuspended()` | Returns `runPolicy.Suspend` |
| `Suspend()` | Sets `runPolicy.Suspend = true` |
| `RunWithPodSetsInfo()` | Unsuspends; iterates replica specs injecting node affinity |
| `PodSets()` | One PodSet per replica type (Master, Worker, PS, etc.) |
| `Finished()` | Checks job status conditions |

## Pod Set Ordering

Replica specs are iterated in a stable, deterministic order (sorted by replica type name) to ensure consistent PodSet ordering between the Workload and the job.

## MultiKueue Support

Implemented by individual adapters where supported (PyTorchJob, TFJob).
