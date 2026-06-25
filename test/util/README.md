# test/util/

Shared test helpers used across integration and e2e test layers.

## Contents

| File | Purpose |
|---|---|
| `factory.go` | Object builder helpers — `MakeJob`, `MakeWorkload`, `MakeClusterQueue`, `MakeLocalQueue`, `MakeResourceFlavor`, etc. Each returns a fluent builder for concise test setup. |
| `util_scheduling.go` | Scheduling-specific helpers — waiting for workload admission, verifying quota usage, asserting queue depths. |

## Usage Pattern

Tests construct objects with the builders and use `gomega.Eventually` + these helpers to assert on controller outcomes:

```go
job := testing.MakeJob("my-job", ns.Name).Queue("lq").Obj()
Expect(k8sClient.Create(ctx, job)).Should(Succeed())
util.ExpectWorkloadToBeAdmittedAs(ctx, k8sClient, wl, admission)
```
