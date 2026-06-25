# test/integration/multikueue/

Integration tests for MultiKueue (multi-cluster workload federation).

## Structure

| Directory | What it covers |
|---|---|
| `scheduler/` | MultiKueue scheduler integration — workload dispatch to worker clusters, admission synchronisation |
| `tas/` | Topology-Aware Scheduling across MultiKueue clusters |

## Setup

The multikueue integration tests simulate a two-cluster environment within a single envtest process. Two separate `envtest.Environment` instances are started (manager cluster + worker cluster). The MultiKueueCluster objects point to in-process kubeconfig secrets rather than real remote clusters.

## Running

```bash
make test-integration-multikueue
# or
go test ./test/integration/multikueue/... -v
```
