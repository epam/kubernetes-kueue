# test/integration/singlecluster/controller/

Integration tests for Kueue's controller layer.

## Sub-packages

| Directory | What it tests |
|---|---|
| `core/` | Core resource controllers: ClusterQueue, LocalQueue, Workload, ResourceFlavor, Cohort, AdmissionCheck reconcilers. Verifies status updates, finalizers, and cross-resource interactions. |
| `jobs/` | Per-job-type adapter integration tests (one package per framework). |
| `admissionchecks/provisioning/` | ProvisioningRequest AdmissionCheck controller — creates and watches ProvisioningRequest objects. |
| `concurrentadmission/` | ConcurrentAdmission feature gate — parallel flavor pursuit across multiple candidate flavors for one workload. |
| `dra/` | Dynamic Resource Allocation — DeviceClass + ResourceClaim integration with Workload PodSets. |
| `failurerecovery/` | Pod termination failure recovery — detects failed pods and re-admits workloads. |
| `jobframework/setup/` | Job framework registration and setup — verifies that all registered adapters start cleanly. |

## jobs/ sub-packages

One package per supported job type. Each verifies: workload creation, status propagation, suspend/resume, preemption, quota accounting, and webhook validation for that specific adapter.

Adapters tested: `job`, `jobset`, `pod`, `raycluster`, `rayjob`, `pytorchjob`, `tfjob`, `jaxjob`, `paddlejob`, `xgboostjob`, `mpijob`, `sparkapplication`, `appwrapper`, `trainjob`, `statefulset`.
