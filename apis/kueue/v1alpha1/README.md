# apis/kueue/v1alpha1/

Alpha-stage CRD types for the `kueue.x-k8s.io` group.

## Types

### `Topology`

Defines the hierarchical node topology used by Topology-Aware Scheduling (TAS).

```go
type TopologySpec struct {
    Levels []TopologyLevel  // Ordered from coarsest to finest: rack → host → ...
}

type TopologyLevel struct {
    NodeLabel string  // Node label key used to group nodes at this level
}
```

**Example:** A two-level topology with `cloud.provider.com/topology-rack` and `kubernetes.io/hostname` allows Kueue to prefer placing pods on the same rack, then the same node.

## Status

All types here are `v1alpha1` and may change between releases without backwards compatibility guarantees.

## Relationship to v1beta1

`Topology` is also referenced from `v1beta1/topology_types.go` via conversion. The TAS controller (`pkg/controller/tas/`) watches `Topology` objects and uses them to make scheduling placement decisions.
