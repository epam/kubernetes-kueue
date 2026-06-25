# test/

Kueue's complete test suite. Three major layers — integration, e2e, and performance — plus shared utilities.

## Structure

| Directory | Layer | Description |
|---|---|---|
| `integration/` | Integration | In-process tests using envtest (real API server, no real nodes) |
| `e2e/` | End-to-end | Tests against a live kind cluster with Kueue deployed |
| `performance/` | Performance | Scheduler throughput benchmarks |
| `util/` | Shared | Cross-layer test helpers and object factories |

## Test Taxonomy

### Integration tests (`test/integration/`)

Run with `make test-integration`. Use `controller-runtime/envtest` which starts a real `kube-apiserver` and `etcd` but no kubelet. Pods never actually schedule onto nodes — workload admission and controller reconciliation are what's verified.

Organised by cluster topology:
- `singlecluster/` — standard single-cluster tests (the vast majority)
- `multikueue/` — manager+worker federation tests

### End-to-end tests (`test/e2e/`)

Run with `make test-e2e`. Require a live kind cluster. Kueue is deployed from the local image. Tests verify the full stack including webhook admission, pod scheduling, and real node resource accounting.

Organised by scope:
- `singlecluster/` — single-cluster job types (baseline core + extended integrations)
- `multikueue/` — multi-cluster federation
- `tas/` — Topology-Aware Scheduling scenarios
- `sequential/` — tests that must not run concurrently (e.g., HA failover)
- `certmanager/` — TLS certificate management integration
- `dra/` — Dynamic Resource Allocation integration
- `upgrade/` — rolling upgrade validation

### Performance tests (`test/performance/`)

Scheduler throughput benchmarks. Run a synthetic workload generator against a kind cluster and measure scheduling latency and throughput at scale.

## Framework

All tests use **Ginkgo v2** (BDD) with **Gomega** matchers. `suite_test.go` files bootstrap the Ginkgo suite and register the scheme.
