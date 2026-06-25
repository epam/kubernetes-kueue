# test/e2e/

End-to-end tests for Kueue. Run against a live kind cluster with Kueue deployed.

## Prerequisites

- `kind` cluster running (`make kind-image-build kind-load`)
- Kueue installed from local image (`make kind-deploy`)
- `KUBECONFIG` pointing at the kind cluster

## Structure

| Directory | Description |
|---|---|
| `singlecluster/` | Core single-cluster scenarios |
| `singlecluster/baseline/` | Standard job types, fair sharing, visibility, kueuectl, TAS basics |
| `singlecluster/extended/` | Extended integrations (JobSet, Ray, KubeFlow, LWS, AppWrapper) |
| `multikueue/` | MultiKueue federation (manager + 2 worker clusters) |
| `multikueue/baseline/` | Basic multi-cluster workload dispatch |
| `multikueue/extended/` | Extended multi-cluster scenarios |
| `multikueue/sequential/` | Sequential multi-cluster tests |
| `multikueue/dra/` | DRA in multi-cluster setup |
| `tas/` | Topology-Aware Scheduling E2E |
| `tas/baseline/` | Core TAS scenarios (Job, Pod group, StatefulSet) |
| `tas/extended/` | Extended TAS (JobSet, LWS, Ray, MPIJob, PyTorch) |
| `sequential/` | Tests requiring sequential execution (cluster-wide config changes) |
| `sequential/baseline/` | HA, metrics, reconciliation, visibility server |
| `sequential/extended/` | SparkApplication, workload identifier annotations |
| `certmanager/` | cert-manager TLS integration |
| `dra/` | DRA E2E baseline |
| `upgrade/` | Rolling upgrade validation |

## Running

```bash
# Full E2E suite
make test-e2e

# TAS E2E only
make test-e2e-tas

# MultiKueue E2E (requires 3 clusters)
make test-e2e-multikueue
```

## Cluster Setup

E2E tests create their own namespaces, ClusterQueues, LocalQueues, and ResourceFlavors per test. `BeforeEach` / `AfterEach` clean up all resources. A `DeferCleanup` in `suite_test.go` tears down shared infrastructure.
