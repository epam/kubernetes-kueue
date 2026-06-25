# cmd/importer/mapping/

Resource mapping logic for the Kueue importer. Determines which ClusterQueue and LocalQueue to assign to imported workloads.

## Mapping Rules

The importer supports several mapping strategies:
1. **Label-based** — read `kueue.x-k8s.io/queue-name` from the Pod, map to a LocalQueue
2. **Namespace-based** — map namespaces to queues via a config file
3. **Default** — all unmatched pods go to a default ClusterQueue

## Configuration

```yaml
# importer-mapping.yaml
mappings:
  - selector:
      matchLabels:
        team: "ml-team"
    localQueue: "ml-queue"
    clusterQueue: "ml-cq"
  - default: true
    localQueue: "default-queue"
    clusterQueue: "default-cq"
```
