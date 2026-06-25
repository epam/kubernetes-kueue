# test/e2e/multikueue/

End-to-end tests for MultiKueue — Kueue's multi-cluster workload federation.

## Cluster Setup

MultiKueue E2E requires three kind clusters:
- **Manager cluster** — runs Kueue with MultiKueue controller enabled
- **Worker cluster 1** — receives dispatched workloads
- **Worker cluster 2** — receives dispatched workloads

Clusters are created by `hack/testing/e2e-multikueue-test.sh`.

## Sub-packages

| Directory | What it tests |
|---|---|
| `baseline/` | Core multi-cluster dispatch: workloads submitted to manager, admitted on workers; `MultiKueueCluster` lifecycle; cross-cluster TAS |
| `extended/` | Extended scenarios: cluster failure handling, worker re-admission, admission check integration |
| `sequential/` | Scenarios that modify cluster-wide config and must run in isolation |
| `dra/` | Dynamic Resource Allocation across MultiKueue clusters |

## Key Objects

- `MultiKueueConfig` — names the worker clusters a MultiKueue ClusterQueue can dispatch to
- `MultiKueueCluster` — points to a worker cluster kubeconfig secret
- `AdmissionCheck` (type `kueue.x-k8s.io/multikueue`) — gates admission until a worker cluster accepts the workload
