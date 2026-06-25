# test/integration/singlecluster/scheduler/podsready/

Integration tests for the WaitForPodsReady feature.

## What's tested

- When `waitForPodsReady.enable=true`, admitted workloads wait for pods to reach Ready before the next workload is admitted
- Timeout: if pods don't become Ready within `waitForPodsReady.timeout`, the workload is evicted and requeued
- `Workload.status.conditions[PodsReady]` transitions
- Interaction with preemption — preempted workloads reset the PodsReady countdown
