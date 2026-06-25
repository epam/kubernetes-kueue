# pkg/podset/

PodSet types and operations. A `PodSet` is Kueue's fundamental unit of resource accounting — a group of pods with identical requirements that must all be admitted together.

## Key Types

### `PodSet` (in `apis/kueue/v1beta1/workload_types.go`)

```go
type PodSet struct {
    Name     string
    Template corev1.PodTemplateSpec
    Count    int32
    // TAS fields
    TopologyRequest *PodSetTopologyRequest
}
```

### `PodSetInfo`

Runtime information passed from the scheduler back to the job adapter:

```go
type PodSetInfo struct {
    Name         string
    Count        int32                          // possibly reduced for partial admission
    NodeSelector map[string]string              // injected node affinity
    Tolerations  []corev1.Toleration            // injected tolerations
    TopologyAssignment *kueue.TopologyAssignment // TAS placement
    Annotations  map[string]string
    Labels       map[string]string
}
```

## Key Functions

- `Merge(psi PodSetInfo, template *corev1.PodTemplateSpec)` — apply `PodSetInfo` (node selectors, tolerations) into a pod template
- `FromAssignment(assignment PodSetAssignment)` — convert scheduler assignment to PodSetInfo for job adapters

## Design Rationale

PodSets abstract away the structural differences between job types. A `batch/Job` has one PodSet; a `PyTorchJob` has two (Master + Worker); a `JobSet` has N. The scheduler treats all of them uniformly.
