# test/integration/singlecluster/controller/core/

Integration tests for Kueue's core resource controllers.

## What's tested

Each core controller is exercised against a real API server (envtest):

| Controller | What's verified |
|---|---|
| `ClusterQueueReconciler` | Status updates (Active/Terminating/Inactive), quota tracking, cohort membership, termination/finalizer cleanup |
| `LocalQueueReconciler` | Status mirroring from ClusterQueue, creation/deletion lifecycle |
| `WorkloadReconciler` | Admission status propagation, finalizer, owner reference cleanup, eviction on quota change |
| `ResourceFlavorReconciler` | Status and taints/tolerations validation |
| `CohortReconciler` | Cohort tree updates, resource accumulation |
| `AdmissionCheckReconciler` | Check lifecycle, active/inactive transitions |

## Cross-resource interactions

Tests also verify multi-resource scenarios:
- Deleting a ClusterQueue while workloads are admitted
- LocalQueue pointing to a non-existent ClusterQueue
- Workload referencing a deleted LocalQueue
- ResourceFlavor referenced by an active ClusterQueue cannot be deleted
