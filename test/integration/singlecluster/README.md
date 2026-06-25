# test/integration/singlecluster/

Integration tests for all single-cluster Kueue functionality.

## Sub-packages

| Directory | What it covers |
|---|---|
| `controller/core/` | Core controllers: ClusterQueue, LocalQueue, Workload, ResourceFlavor, Cohort, AdmissionCheck reconcilers |
| `controller/jobs/` | Per-adapter integration tests for each job type (batch/Job, JobSet, Pod, Ray*, KubeFlow*, etc.) |
| `controller/admissionchecks/provisioning/` | ProvisioningRequest AdmissionCheck controller |
| `controller/concurrentadmission/` | ConcurrentAdmission feature |
| `controller/dra/` | Dynamic Resource Allocation integration |
| `controller/failurerecovery/` | Pod termination failure recovery controller |
| `controller/jobframework/setup/` | Job framework setup and registration |
| `scheduler/` | Scheduler integration tests (main suite + sub-suites below) |
| `scheduler/fairsharing/` | Fair sharing and DRS-based preemption |
| `scheduler/delayedadmission/` | Delayed admission / WaitForPodsReady integration |
| `scheduler/inadmissible/` | Inadmissible workload requeueing |
| `scheduler/podsready/` | WaitForPodsReady-specific scenarios |
| `scheduler/quotacheckstrategy/` | Quota check strategy variants |
| `scheduler/resourcetransformations/` | Resource transformation rules |
| `scheduler/excluderesources/` | ExcludeResourcePrefixes scenarios |
| `tas/` | Topology-Aware Scheduling |
| `webhook/core/` | Core webhook validation/mutation tests |
| `webhook/jobs/` | Job adapter webhook tests |
| `importer/` | Kueue importer tool integration tests |
| `kueuectl/` | kueuectl command integration tests (against a live API server) |
| `conversion/` | CRD conversion webhook tests |
