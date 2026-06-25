# test/integration/singlecluster/controller/jobs/

Per-adapter integration tests. Each sub-package tests one job type's Kueue adapter end-to-end using envtest.

## Coverage per package

Each package verifies the full adapter lifecycle:
1. Job created → Workload created automatically
2. Workload admitted → Job unsuspended
3. Job completes → Workload marked Finished
4. Workload preempted → Job suspended
5. Quota accounting correct (resource requests reflected in ClusterQueue usage)
6. Webhook validation/mutation for the job type

## Sub-packages

| Package | Job type |
|---|---|
| `job/` | `batch/v1 Job` |
| `jobset/` | `jobset.x-k8s.io/v1alpha2 JobSet` |
| `pod/` | `v1 Pod` (pod group + single pod) |
| `statefulset/` | `apps/v1 StatefulSet` |
| `raycluster/` | `ray.io/v1 RayCluster` |
| `rayjob/` | `ray.io/v1 RayJob` |
| `pytorchjob/` | `kubeflow.org/v1 PyTorchJob` |
| `tfjob/` | `kubeflow.org/v1 TFJob` |
| `jaxjob/` | `kubeflow.org/v1 JAXJob` |
| `paddlejob/` | `kubeflow.org/v1 PaddleJob` |
| `xgboostjob/` | `kubeflow.org/v1 XGBoostJob` |
| `mpijob/` | `kubeflow.org/v2beta1 MPIJob` |
| `sparkapplication/` | `sparkoperator.k8s.io/v1beta2 SparkApplication` |
| `appwrapper/` | `workload.codeflare.dev/v1beta2 AppWrapper` |
| `trainjob/` | `trainer.kubeflow.org/v1alpha1 TrainJob` |
