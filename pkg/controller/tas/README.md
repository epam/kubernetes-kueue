# pkg/controller/tas/

Topology-Aware Scheduling (TAS) controller. Ensures workloads are placed on nodes that satisfy topology constraints (rack, host, zone) to minimize cross-rack or cross-host communication for distributed jobs.

## Concept

TAS allows a workload to declare that all its pods (or subsets) must be placed within the same topology domain:

```yaml
# In PodSet:
topologyRequest:
  required:
    level: "cloud.provider.com/topology-rack"
```

This means all pods of this PodSet must run on nodes in the same rack.

## How It Works

1. **Topology discovery** — the TAS controller watches `Node` objects and maps them to topology levels defined in `Topology` CRDs
2. **Domain calculation** — builds a map of topology domains (e.g., `rack-1 → [node-a, node-b, node-c]`)
3. **Placement assignment** — during scheduling (FlavorAssigner), TAS selects a topology domain that can accommodate all pods, injecting `nodeAffinity` rules into the pod template
4. **Controller sync** — this controller keeps the topology domain map up to date as nodes join/leave

## Key Types

### `TASController`

Watches `Node` objects and maintains the topology domain cache used by the scheduler.

### `TopologyAssignment` (in `apis/kueue/v1beta1/workload_types.go`)

```go
type TopologyAssignment struct {
    Levels  []string                  // topology hierarchy levels used
    Domains []TopologyDomainAssignment // domain → pod count mapping
}
```

Written to `Workload.status.admission.podSetAssignments[].topologyAssignment` when admitted.

## Sub-packages

| Package | Purpose |
|---|---|
| [`indexer/`](indexer/) | Node topology indexes for fast domain lookup |

## Feature Gate

`features.TopologyAwareScheduling` — when disabled, TAS fields are ignored.

## Supported Levels

Any node label can be a topology level. Common examples:
- `cloud.provider.com/topology-rack`
- `kubernetes.io/hostname`
- `topology.kubernetes.io/zone`
