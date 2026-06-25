# test/integration/singlecluster/scheduler/resourcetransformations/

Integration tests for resource transformation rules.

## What's tested

Resource transformations allow ClusterQueues to rewrite resource requests before quota accounting. For example, a transformation can map `vendor.com/gpu: 1` to `cpu: 4` for accounting purposes.

Tests verify:
- Transformed resource names appear in ClusterQueue usage, not the original names
- Workloads using extended resources are correctly accounted after transformation
- Transformations interact correctly with borrowing and preemption
- Invalid transformation configurations are rejected by the webhook
