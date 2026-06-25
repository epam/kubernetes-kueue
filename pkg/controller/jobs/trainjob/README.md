# pkg/controller/jobs/trainjob/

Integration adapter for `TrainJob` (`trainer.kubeflow.org/v1alpha1`). `TrainJob` is the next-generation KubeFlow Training Operator API that replaces individual PyTorchJob/TFJob/etc. with a unified training API.

## GenericJob Implementation

| Method | Behavior |
|---|---|
| `IsSuspended()` | Checks `trainjob.Spec.Suspend` |
| `Suspend()` | Sets `Spec.Suspend = true` |
| `RunWithPodSetsInfo()` | Unsuspends; injects node affinity into trainer pod template |
| `PodSets()` | Extracts PodSets from the trainer runtime spec |
| `Finished()` | Checks `TrainJob.Status.Conditions` for completion |

## Relationship to KubeFlow Training Operator

`TrainJob` creates a `JobSet` internally. The KubeFlow Training Operator v2 manages the mapping from `TrainJob` → `JobSet` → individual `Job`s. Kueue sees the `TrainJob` as the top-level object.

## MultiKueue Support

Full MultiKueue support — dispatches the entire TrainJob to a worker cluster.
