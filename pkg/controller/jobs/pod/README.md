# pkg/controller/jobs/pod/

Integration adapter for plain Kubernetes `Pod`s managed by Kueue. Supports both individual pods and pod groups (multiple pods that must be co-admitted).

## Pod Group Concept

A "pod group" is a set of pods with the same `kueue.x-k8s.io/pod-group-name` label. Kueue treats them as a single workload and only admits the group when all pods in the group can be scheduled together.

```yaml
metadata:
  labels:
    kueue.x-k8s.io/pod-group-name: "my-training-job"
    kueue.x-k8s.io/pod-group-total-count: "4"
    kueue.x-k8s.io/queue-name: "my-queue"
```

## GenericJob Implementation

| Method | Behavior |
|---|---|
| `IsSuspended()` | Checks if pod has `schedulingGates` containing Kueue's gate |
| `Suspend()` | Adds scheduling gate to prevent pod scheduling |
| `RunWithPodSetsInfo()` | Removes scheduling gate; injects node affinity |
| `PodSets()` | One PodSet per role group (grouped by pod template hash) |
| `Finished()` | True when all pods in group are `Succeeded` or any is `Failed` |

## Scheduling Gate

Kueue gates unscheduled pods via `schedulingGates`:
```yaml
spec:
  schedulingGates:
  - name: "kueue.x-k8s.io/admission"
```

The pod stays in `Pending` state until Kueue removes this gate.

## Constants (`constants/`)

Defines label keys and annotation keys specific to the pod integration.

## MultiKueue Support

Not supported — pods cannot be meaningfully dispatched to remote clusters.

## Deployment / StatefulSet Pods

Pods created by `Deployment` or `StatefulSet` controllers can also use this integration when `kueue.x-k8s.io/managed: "true"` label is set on the pod. Each pod gets its own Workload.
