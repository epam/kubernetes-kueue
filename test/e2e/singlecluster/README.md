# test/e2e/singlecluster/

End-to-end tests for single-cluster Kueue deployments.

## Sub-packages

### `baseline/`

Core Kueue functionality tested against a real cluster:
- `job_test.go` — batch/Job admit, suspend, resume, preemption
- `pod_test.go` — plain Pod + pod group admission
- `deployment_test.go` — Deployment workload management
- `statefulset_test.go` — StatefulSet workload management
- `fair_sharing_test.go` — DRS-based fair sharing and preemption
- `tas_test.go` — Topology-Aware Scheduling basics
- `visibility_test.go` — VisibilityOnDemand API (pending workloads)
- `kueuectl_test.go` — kueuectl CLI commands against a live cluster
- `metrics_test.go` — Prometheus metrics endpoint validation
- `prometheus_test.go` — Prometheus scrape integration
- `concurrent_admission_test.go` — ConcurrentAdmission feature
- `e2e_v1beta1_test.go` — v1beta1 API surface validation

### `extended/`

Third-party job framework integrations:
- `jobset_test.go` — JobSet admission and lifecycle
- `kuberay_test.go` — RayJob / RayCluster admission
- `pytorchjob_test.go` — PyTorchJob admission
- `jaxjob_test.go` — JAXJob admission
- `leaderworkerset_test.go` — LeaderWorkerSet admission
- `appwrapper_test.go` — AppWrapper admission
- `trainjob_test.go` — TrainJob admission
