# pkg/controller/jobs/jobset/

Integration adapter for `JobSet` (`jobset.x-k8s.io/v1alpha2`). A JobSet is a collection of homogeneous Jobs with coordinated lifecycle management — ideal for distributed training with multiple worker roles.

## GenericJob Implementation

| Method | Behavior |
|---|---|
| `IsSuspended()` | Returns `jobset.Spec.Suspend` |
| `Suspend()` | Sets `Suspend = true` on all ReplicatedJobs |
| `RunWithPodSetsInfo()` | Unsuspends; injects node affinity per ReplicatedJob role |
| `PodSets()` | Returns one PodSet per `ReplicatedJob` in the JobSet |

## Pod Set Mapping

Each `ReplicatedJob` in the `JobSet` spec maps to one PodSet:
```
JobSet.spec.replicatedJobs:
  - name: "leader"    → PodSet { name: "leader", count: 1 }
  - name: "worker"    → PodSet { name: "worker", count: N }
```

## MultiKueue Support

Full MultiKueue support. The manager creates a JobSet copy on the worker cluster and syncs status back.

## Key Behaviors

- When suspended, all child Jobs are also suspended (propagated by the JobSet controller)
- Kueue waits for all pods across all ReplicatedJobs to be ready before setting `PodsReady`
- Failure of any sub-job propagates as JobSet failure
