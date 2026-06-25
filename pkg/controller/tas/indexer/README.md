# pkg/controller/tas/indexer/

Field indexes for Topology-Aware Scheduling. Enables efficient lookups of nodes by their topology labels.

## Indexes

- **Node by topology label** — for each topology level label, indexes nodes by their label value (e.g., all nodes in `rack-1`)
- **ResourceFlavor by topology** — enables fast lookup of which flavors reference a given Topology CRD

## Usage

The TAS controller uses these indexes to quickly find which nodes belong to each topology domain without scanning all nodes on every scheduling cycle.

```go
// Find all nodes in topology domain "rack-1":
nodeList := &corev1.NodeList{}
client.List(ctx, nodeList, client.MatchingFields{
    indexer.NodeByTopologyLabel("cloud.provider.com/topology-rack"): "rack-1",
})
```
