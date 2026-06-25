# pkg/controller/jobs/raycluster/

Integration adapter for `RayCluster` (`ray.io/v1`). Manages Ray distributed computing clusters through Kueue.

## GenericJob Implementation

| Method | Behavior |
|---|---|
| `IsSuspended()` | Checks `raycluster.Spec.Suspend` |
| `Suspend()` | Sets `Spec.Suspend = true` |
| `RunWithPodSetsInfo()` | Unsuspends; injects node affinity into head + worker group templates |
| `PodSets()` | One PodSet per worker group + one for the head node |

## Pod Set Mapping

```
RayCluster.spec.headGroupSpec        → PodSet { name: "head", count: 1 }
RayCluster.spec.workerGroupSpecs[0]  → PodSet { name: "worker-0", count: N }
RayCluster.spec.workerGroupSpecs[1]  → PodSet { name: "worker-1", count: M }
```

## MultiKueue Support

Full MultiKueue support via `MultiKueueAdapter`. Dispatches RayCluster to worker clusters.

## Notes

- Ray Autoscaler integration: if autoscaler is enabled on the RayCluster, Kueue manages initial quota but autoscaler can grow within bounds
- The head node is always scheduled first; workers are added once the head is ready
