# pkg/util/testing/

Test builder utilities and Gomega matchers for Kueue integration and unit tests.

## Builder Pattern

All builders use a fluent API:
```go
wl := utiltesting.MakeWorkload("my-wl", "default").
    Queue("my-queue").
    Request(corev1.ResourceCPU, "1").
    Priority(100).
    Obj()
```

### Key Builders

| Builder | Creates |
|---|---|
| `MakeWorkload(name, ns)` | `kueue.Workload` |
| `MakeClusterQueue(name)` | `kueue.ClusterQueue` |
| `MakeLocalQueue(name, ns)` | `kueue.LocalQueue` |
| `MakeResourceFlavor(name)` | `kueue.ResourceFlavor` |
| `MakeCohort(name)` | `kueue.Cohort` |
| `MakeAdmissionCheck(name)` | `kueue.AdmissionCheck` |
| `MakePodSet(name)` | `kueue.PodSet` |
| `MakeTopology(name)` | `kueuev1alpha1.Topology` |

## Gomega Matchers

```go
// In test:
gomega.Eventually(func() bool {
    return gomega.HaveConditionStatus(
        kueue.WorkloadAdmitted, metav1.ConditionTrue,
    ).Match(wl)
})
```

## Sub-packages

| Package | Purpose |
|---|---|
| [`metrics/`](metrics/) | Gomega matchers for Prometheus metrics |
| [`v1beta1/`](v1beta1/) | v1beta1 API-specific builders |
| [`v1beta2/`](v1beta2/) | v1beta2 API-specific builders |
