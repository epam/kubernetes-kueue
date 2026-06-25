# test/integration/singlecluster/scheduler/excluderesources/

Integration tests for the `ExcludeResourcePrefixes` feature.

## What's tested

`managerConfig.integrations.podOptions.podResourcesExcludeList` allows operators to define resource name prefixes that Kueue should ignore when computing workload resource requests.

Tests verify:
- Resources matching an excluded prefix are not counted toward quota
- Workloads with only excluded resources are admitted without any quota check
- Excluded resources do not appear in ClusterQueue usage metrics
- Non-excluded resources on the same workload are still accounted correctly
