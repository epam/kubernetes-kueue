# pkg/controller/jobs/deployment/

Integration adapter for Kubernetes `Deployment`. Allows Kueue to manage resource quota for Deployments — each pod replica gets its own Workload.

## Model

Unlike batch jobs, Deployments run continuously. Kueue creates one `Workload` per pod (not per Deployment). The Workload is tied to the pod's lifecycle.

## GenericJob Implementation

| Method | Behavior |
|---|---|
| `IsSuspended()` | Returns `deployment.Spec.Replicas == 0` |
| `Suspend()` | Scales Deployment to 0 replicas |
| `RunWithPodSetsInfo()` | Scales back to desired replicas; injects node affinity |
| `PodSets()` | Single PodSet for the Deployment pod template |

## Use Case

Useful for long-running services that consume GPU or special hardware resources that need quota tracking. Less common than batch job integration.

## Limitations

No MultiKueue support. Deployment preemption scales the deployment to 0, which may be disruptive for production services.
