# pkg/controller/jobs/rayjob/

Integration adapter for `RayJob` (`ray.io/v1`). A `RayJob` creates a `RayCluster` internally and submits a job to it.

## GenericJob Implementation

| Method | Behavior |
|---|---|
| `IsSuspended()` | Checks `rayjob.Spec.Suspend` |
| `Suspend()` | Sets `Spec.Suspend = true` |
| `RunWithPodSetsInfo()` | Unsuspends; injects node affinity into embedded cluster template |
| `PodSets()` | Delegates to the embedded RayCluster spec |
| `Finished()` | Returns true when `rayjob.Status.JobStatus` is `SUCCEEDED` or `FAILED` |

## Relationship to RayCluster

`RayJob` is a higher-level abstraction that:
1. Creates a `RayCluster` when submitted
2. Submits a Python script/entrypoint to that cluster
3. Deletes the cluster when the job finishes (configurable)

Kueue manages the `RayJob`, which in turn controls the lifecycle of the underlying `RayCluster`.

## MultiKueue Support

Full MultiKueue support. The RayJob is dispatched to a worker cluster, including its embedded cluster template.
