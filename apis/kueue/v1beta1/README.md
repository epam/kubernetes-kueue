# apis/kueue/v1beta1/

The primary API version for all Kueue CRDs. This is the **hub (storage) version** — all other versions convert through v1beta1 for storage in etcd.

## Core Types

### `ClusterQueue`

Cluster-scoped resource quota pool with scheduling policies.

**Key spec fields:**
- `resourceGroups` — list of `ResourceGroup`, each naming resources and the flavors that provide them
- `cohort` — optional name linking this CQ to a borrowing group
- `queueingStrategy` — `BestEffortFIFO` or `StrictFIFO`
- `preemption` — when/how to evict lower-priority workloads
- `admissionChecks` / `admissionChecksStrategy` — external gates before admission
- `stopPolicy` — `None`, `Hold`, `HoldAndDrain`
- `fairSharing` — per-CQ fair sharing weight

**Key status fields:**
- `conditions` — `Active`, `Pending`
- `admittedWorkloads`, `pendingWorkloads`, `reservingWorkloads`
- `flavorsUsage` — current resource consumption per flavor

### `LocalQueue`

Namespace-scoped entry point for submitting workloads to a `ClusterQueue`.

```go
type LocalQueueSpec struct {
    ClusterQueue ClusterQueueReference
    StopPolicy   StopPolicy
}
```

### `Workload`

Kueue's internal representation of any batch job. Created automatically by job adapters.

**Key spec fields:**
- `podSets` — list of `PodSet` (name + template + count + topology request)
- `queueName` — which LocalQueue to use
- `priorityClassName` / `priority` — scheduling priority
- `maximumExecutionTime` — deadline from first admission

**Key status fields:**
- `admission` — assigned flavors and node affinities (set when admitted)
- `conditions` — `Admitted`, `Finished`, `Evicted`, `PodsReady`, `QuotaReserved`
- `reclaimablePods` — pods that can be released without fully evicting the workload

### `ResourceFlavor`

Maps a named flavor to node characteristics.

```go
type ResourceFlavorSpec struct {
    NodeLabels     map[string]string  // Node affinity to inject into pods
    NodeTaints     []corev1.Taint     // Tolerations to inject
    Tolerations    []corev1.Toleration
    TopologyName   *string            // Reference to a Topology CRD for TAS
}
```

### `Cohort`

Named group of ClusterQueues that can borrow quota from each other. Cohorts can be hierarchical (parent-child).

### `AdmissionCheck`

Pluggable pre-admission gate. Controllers implement the check (e.g., MultiKueue, ProvisioningRequest). Workloads wait until all required checks are `Ready`.

### `WorkloadPriorityClass`

Extends `PriorityClass` with Kueue-specific priority semantics. Referenced in `Workload.spec.priorityClassName`.

### `MultiKueueConfig` / `MultiKueueCluster`

Configuration for MultiKueue federation. `MultiKueueCluster` holds the kubeconfig secret reference for a single worker cluster.

## Constants (`constants.go`)

Key labels and annotations:
- `kueue.x-k8s.io/queue-name` — label on jobs to route them to a LocalQueue
- `kueue.x-k8s.io/job-uid` — label on Workloads for fast reverse lookup
- `kueue.x-k8s.io/managed` — label enabling Kueue management of plain pods

## Conversion

Files named `*_conversion.go` implement conversion from/to `v1beta2`. `zz_generated.conversion.go` handles boilerplate. Hub version has no conversion functions (it IS the hub).
