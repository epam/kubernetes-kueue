# pkg/controller/jobs/mpijob/

Integration adapter for `MPIJob` (`kubeflow.org/v2beta1`). Manages MPI (Message Passing Interface) distributed training jobs.

## GenericJob Implementation

| Method | Behavior |
|---|---|
| `IsSuspended()` | Checks `mpijob.Spec.RunPolicy.Suspend` |
| `Suspend()` | Sets `RunPolicy.Suspend = true` |
| `RunWithPodSetsInfo()` | Unsuspends; injects node affinity into launcher + worker templates |
| `PodSets()` | Two PodSets: `launcher` (×1) and `worker` (×N) |

## Pod Set Mapping

```
MPIJob.spec.mpiReplicaSpecs:
  Launcher: { replicas: 1, template: {...} }  → PodSet "launcher"
  Worker:   { replicas: N, template: {...} }  → PodSet "worker"
```

## MultiKueue Support

Not supported.

## Notes

- MPI Launcher is the coordinator; Workers execute the distributed computation
- Kueue suspends by setting `RunPolicy.Suspend` — the MPI operator propagates this to all pods
