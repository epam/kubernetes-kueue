# pkg/util/testing/metrics/

Gomega matchers for asserting Prometheus metric values in tests.

## Key Functions

- `ExpectAdmittedWorkloadsTotalMetric(cq string, n int)` — assert admitted workloads count
- `ExpectPendingWorkloadsMetric(cq string, active, inadmissible int)` — assert pending counts
- `ExpectEvictedWorkloadsTotalMetric(cq, reason string, n int)` — assert eviction counts

## Usage

```go
// In integration test:
gomega.Expect(metrics.ExpectAdmittedWorkloadsTotalMetric("my-cq", 3)).To(gomega.Succeed())
```

These matchers read from the Prometheus registry directly (no HTTP call needed), making them fast in integration tests.
