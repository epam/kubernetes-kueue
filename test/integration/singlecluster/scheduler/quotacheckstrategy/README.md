# test/integration/singlecluster/scheduler/quotacheckstrategy/

Integration tests for quota check strategy variants.

## What's tested

Kueue supports two quota check strategies (controlled by `managerConfig.integrations.podOptions.namespaceSelector` and `integrations.jobSet.managedByLabel`):

- **Strict**: workload is not admitted if any resource exceeds the ClusterQueue quota, even within a cohort
- **Standard** (default): borrowing from the cohort is permitted up to the cohort's total quota

Tests verify the boundary conditions and interactions with borrowing and cohort hierarchies.
