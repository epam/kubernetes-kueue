# pkg/controller/jobs/rayservice/

Integration adapter for `RayService` (`ray.io/v1`). A `RayService` is a long-running Ray cluster that serves online inference requests.

## GenericJob Implementation

| Method | Behavior |
|---|---|
| `IsSuspended()` | Checks `rayservice.Spec.Suspend` |
| `Suspend()` | Sets `Spec.Suspend = true` |
| `RunWithPodSetsInfo()` | Unsuspends; injects node affinity |
| `PodSets()` | PodSets from the embedded RayCluster spec |
| `Finished()` | Always returns false (services run indefinitely) |

## Differences from RayJob

- `RayService` runs continuously (serving inference)
- It manages rolling upgrades of the Ray cluster
- Kueue manages quota for the cluster but cannot "finish" it — only suspend/unsuspend
- The `activeServiceStatus` and `pendingServiceStatus` track active vs. upgrading clusters

## MultiKueue Support

Not supported — online serving services are not typically dispatched to worker clusters.
