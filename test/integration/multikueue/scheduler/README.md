# test/integration/multikueue/scheduler/

Integration tests for the MultiKueue scheduler.

## Purpose

Verifies workload dispatch from the manager cluster to worker clusters, and the full admission lifecycle in a simulated two-cluster environment.

## What's tested

- Workloads submitted to the manager cluster are mirrored on a selected worker cluster
- Worker cluster admission state is synchronised back to the manager cluster
- `MultiKueueCluster` status reflects worker cluster connectivity
- Worker cluster failure: workload is retried on another available worker
- AdmissionCheck gate is released only after the worker cluster admits the workload
- Workload completion on the worker cluster finalises the workload on the manager cluster
