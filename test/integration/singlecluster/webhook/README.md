# test/integration/singlecluster/webhook/

Integration tests for Kueue webhook validation and mutation.

## Sub-packages

| Package | What it tests |
|---|---|
| `core/` | Webhooks for core Kueue resources: ClusterQueue, LocalQueue, Workload, ResourceFlavor, Cohort, AdmissionCheck, WorkloadPriorityClass — validates that invalid specs are rejected and defaulting mutations are applied |
| `jobs/` | Webhooks for all supported job types — validates that `kueue.x-k8s.io/queue-name` label handling, PodSet extraction, and framework-specific field validation work correctly |

## What's tested

- Invalid resources are rejected with descriptive error messages
- Defaulting webhooks set required labels and annotations
- Immutable fields cannot be changed after creation
- Cross-resource references (e.g., LocalQueue → ClusterQueue) are validated at admission time
