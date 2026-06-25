# pkg/controller/jobs/statefulset/

Integration adapter for Kubernetes `StatefulSet`. Similar to the Deployment adapter — manages resource quota for StatefulSet replicas.

## Model

Each pod in the StatefulSet gets its own Kueue `Workload`. The adapter controls resource access by managing StatefulSet replica count.

## GenericJob Implementation

| Method | Behavior |
|---|---|
| `IsSuspended()` | Returns `statefulset.Spec.Replicas == 0` |
| `Suspend()` | Scales StatefulSet to 0 replicas |
| `RunWithPodSetsInfo()` | Scales back to desired; injects node affinity into pod template |
| `PodSets()` | Single PodSet for the StatefulSet pod template |

## Differences from Deployment

StatefulSet pods have stable identities and persistent volume claims. Suspending (scaling to 0) preserves PVCs but may affect stateful workloads. Use with caution for stateful services that depend on stable network identity.

## MultiKueue Support

Not supported.
