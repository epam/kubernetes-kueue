# apis/kueue/

Core Kueue CRD type definitions under the `kueue.x-k8s.io` API group.

## Versions

| Version | Status | Role |
|---|---|---|
| `v1alpha1` | Alpha | `Topology` CRD only |
| `v1beta1` | Stable (hub) | All primary CRDs; this is the storage version |
| `v1beta2` | Beta | Scheduler-internal version; converted from v1beta1 |

## CRDs Defined

| CRD | Version | Description |
|---|---|---|
| `ClusterQueue` | v1beta1 | Cluster-scoped resource quota and scheduling policy |
| `LocalQueue` | v1beta1 | Namespace-scoped queue pointing to a ClusterQueue |
| `Workload` | v1beta1 | Kueue's abstraction over any batch job |
| `ResourceFlavor` | v1beta1 | Node targeting definition (GPU type, zone, etc.) |
| `Cohort` | v1beta1 | Named group of ClusterQueues for borrowing |
| `AdmissionCheck` | v1beta1 | Pluggable external gate before a workload runs |
| `MultiKueueConfig` | v1beta1 | List of worker clusters for MultiKueue |
| `MultiKueueCluster` | v1beta1 | Single worker cluster connection + secret |
| `WorkloadPriorityClass` | v1beta1/v1beta2 | Priority class for workloads |
| `ProvisioningRequestConfig` | v1beta1 | Config for Cluster Autoscaler ProvisioningRequests |
| `Topology` | v1alpha1 | Node topology levels for TAS |

## Key Design Pattern

All CRDs follow the standard Kubernetes pattern:
- `Spec` — desired state (user-editable)
- `Status` — observed state (controller-updated via status subresource)
- `Conditions` — structured status conditions (Active, Admitted, Evicted, etc.)
