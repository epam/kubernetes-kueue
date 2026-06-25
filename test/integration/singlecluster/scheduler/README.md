# test/integration/singlecluster/scheduler/

Integration tests for the Kueue scheduler.

## Main Suite

`suite_test.go` + `workload_controller_test.go` — core scheduling scenarios: quota enforcement, borrowing, preemption, cohort behaviour, workload ordering.

## Sub-suites

Each sub-suite starts its own envtest environment with a specific configuration to isolate feature interactions.

| Directory | Configuration | What it verifies |
|---|---|---|
| `fairsharing/` | FairSharing + DRS enabled | Dominant Resource Sharing preemption, AFS ordering, DRA fair sharing |
| `delayedadmission/` | WaitForPodsReady enabled | Admission held until pods become Ready, timeout eviction |
| `inadmissible/` | Standard | Requeueing of workloads that cannot be admitted (missing flavors, over-quota) |
| `podsready/` | WaitForPodsReady enabled | PodsReady condition transitions |
| `quotacheckstrategy/` | QuotaCheckStrategy variants | Strict vs. relaxed quota check strategies |
| `resourcetransformations/` | ResourceTransformations enabled | Resource request rewriting rules |
| `excluderesources/` | ExcludeResourcePrefixes set | Prefixes excluded from resource accounting |

## Running a single sub-suite

```bash
go test ./test/integration/singlecluster/scheduler/fairsharing/... -v --ginkgo.v
```
