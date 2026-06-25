# pkg/util/tas/

Topology-Aware Scheduling utility helpers.

## Key Functions

- `ValidateTopologyRequest(req *kueue.PodSetTopologyRequest) error` — validate TAS fields on a PodSet
- `TopologyKey(levels []string, domain string) string` — compute a stable key for a topology domain
- `NodeSelectorForDomain(topology *Topology, domain TopologyDomain) corev1.NodeSelector` — build node selector for a specific topology domain

## Usage

Used by the TAS controller and flavor assigner to:
1. Validate that topology requests reference real topology levels
2. Convert topology domain assignments to node selectors that are injected into pod templates
