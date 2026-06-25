# pkg/controller/jobs/job/

Integration adapter for Kubernetes `batch/v1 Job`. This is the primary and most complete job adapter — other adapters reference it as a pattern.

## GenericJob Implementation

| Method | Behavior |
|---|---|
| `IsSuspended()` | Returns `job.Spec.Suspend` |
| `Suspend()` | Sets `job.Spec.Suspend = true` |
| `RunWithPodSetsInfo()` | Sets `Suspend = false`, injects node affinity + pod count per PodSetInfo |
| `PodSets()` | Returns a single PodSet: `{name: "main", count: *job.Spec.Parallelism, template: job.Spec.Template}` |
| `Finished()` | Returns true when `job.Status.CompletionTime != nil` or `Failed` condition is set |
| `PodsReady()` | Returns true when `job.Status.Ready >= *job.Spec.Parallelism` |

## Pod Set Mapping

`batch/v1 Job` has a single pod template → single PodSet named `"main"`.

The `Parallelism` field controls pod count. When Kueue admits with partial admission, `Parallelism` is reduced.

## MultiKueue Support

`MultiKueueAdapter` is implemented. On admission:
1. Manager cluster creates a copy of the Job on the worker cluster
2. Worker runs the Job
3. Manager syncs Job status from worker
4. On completion, manager Workload is marked Finished

## Reclaimable Pods

Implements `JobWithReclaimablePods`. When pods complete successfully (e.g., in a parallel job), their slots can be marked reclaimable — reducing the effective quota held by the workload without evicting it.

## Webhook

- **Mutating**: Injects `kueue.x-k8s.io/job-uid` label, sets default queue if `LocalQueueDefaulting` gate is enabled
- **Validating**: Rejects queue changes while the job is running; validates resource requests
