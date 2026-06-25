# pkg/controller/core/indexer/

Kubernetes informer field indexers for fast lookups across Kueue CRDs. Without indexes, listing objects with specific field values requires a full scan of all objects in the cache.

## Purpose

Controllers frequently need to answer questions like:
- "Which Workloads reference LocalQueue X?"
- "Which LocalQueues reference ClusterQueue Y?"
- "Which Workloads are in namespace Z with queue name Q?"

Without indexes, answering these requires iterating all objects. With indexes, the informer cache maintains precomputed reverse mappings.

## Indexes Defined

| Index | Object | Field | Use Case |
|---|---|---|---|
| `workload/queue` | `Workload` | `spec.queueName` | List workloads by queue |
| `workload/clusterQueue` | `Workload` | admission CQ | List workloads admitted to a CQ |
| `localQueue/clusterQueue` | `LocalQueue` | `spec.clusterQueueName` | Find LQs for a CQ |
| `workload/jobUID` | `Workload` | job owner UID | Fast workload→job lookup |
| `pod/workload` | `Pod` | workload label | Find pods for a workload |

## Usage

```go
// In a controller setup:
if err := indexer.SetupIndexes(ctx, mgr.GetFieldIndexer()); err != nil {
    return err
}

// In reconcile:
var wls kueue.WorkloadList
mgr.GetClient().List(ctx, &wls,
    client.MatchingFields{indexer.WorkloadQueueKey: queueName})
```

## Registration

Indexes must be registered before controllers start (in the `main.go` setup phase) because they cannot be added after informers have started.
