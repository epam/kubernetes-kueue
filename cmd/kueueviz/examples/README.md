# cmd/kueueviz/examples/

Sample Kubernetes manifests demonstrating a typical Kueue setup compatible with the KueueViz dashboard.

## Files

| File | What it creates |
|---|---|
| `00-resource-flavor.yaml` | A `ResourceFlavor` representing a GPU node pool |
| `01-cluster-queues.yaml` | Two `ClusterQueue` objects (high-priority and normal) with resource quotas |
| `01.5-cohorts.yaml` | A `Cohort` grouping the ClusterQueues for borrowing |
| `02-local-queues.yaml` | `LocalQueue` objects in two namespaces pointing to the ClusterQueues |
| `03-agi-job.yaml` — `05-cancer-cure-research.yaml` | Sample `batch/v1 Job` workloads with different priorities and resource requests |
| `07-workload-priority-classes.yaml` | `WorkloadPriorityClass` objects for priority classification |

## Usage

```bash
kubectl apply -f cmd/kueueviz/examples/
```

These manifests create a realistic multi-tenant queue setup that you can visualise immediately in KueueViz.
