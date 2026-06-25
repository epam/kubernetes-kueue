# test/e2e/tas/

End-to-end tests for Topology-Aware Scheduling (TAS).

## What TAS Tests Verify

- `Topology` CRD objects describe node topology (rack, zone, node)
- `ResourceFlavor.topologyName` links a flavor to a topology
- `Workload.podSets[].topologyRequest` specifies placement constraints
- Kueue selects nodes that satisfy the topology constraint before admitting
- Hot-swap (migrating an admitted workload to a different topology node set)

## Sub-packages

### `baseline/`

Core TAS with standard Kubernetes job types:
- `job_test.go` — batch/Job with topology constraints
- `pod_group_test.go` — Pod groups with topology constraints
- `statefulset_test.go` — StatefulSet with topology constraints
- `hotswap_test.go` — Topology hot-swap scenarios

### `extended/`

TAS with third-party job frameworks:
- `jobset_test.go` — JobSet TAS
- `leaderworkerset_test.go` — LeaderWorkerSet TAS
- `rayjob_test.go` — RayJob TAS
- `mpijob_test.go` — MPIJob TAS
- `pytorch_test.go` — PyTorchJob TAS
- `appwrapper_test.go` — AppWrapper TAS
- `trainjob_test.go` — TrainJob TAS
