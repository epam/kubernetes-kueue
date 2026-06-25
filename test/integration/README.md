# test/integration/

Integration tests for Kueue. Run with `make test-integration`.

## Framework

Uses `sigs.k8s.io/controller-runtime/pkg/envtest` which starts a real `kube-apiserver` and `etcd` process (binaries in `$KUBEBUILDER_ASSETS`) but no kubelet. Controllers run in-process. Webhooks run in-process too, served on localhost with a test CA.

All tests use **Ginkgo v2** + **Gomega**. Each sub-package has a `suite_test.go` that:
1. Registers the scheme (all Kueue CRDs + upstream job types).
2. Starts envtest.
3. Creates the shared `k8sClient` and `cfg`.
4. Starts the controllers under test.
5. Registers a `DeferCleanup` to stop the environment.

## Sub-trees

| Directory | What it covers |
|---|---|
| `singlecluster/` | Everything in a single cluster — core controllers, job adapters, scheduler, webhooks, importer, kueuectl, TAS |
| `multikueue/` | Manager+worker federation — MultiKueue controller, TAS across clusters |

## Running

```bash
# All integration tests
make test-integration

# Single package
go test ./test/integration/singlecluster/scheduler/... -v --ginkgo.v

# With race detector
go test -race ./test/integration/singlecluster/controller/core/...
```

Binaries are downloaded automatically by `make envtest`.
