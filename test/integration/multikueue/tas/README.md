# test/integration/multikueue/tas/

Integration tests for Topology-Aware Scheduling (TAS) across MultiKueue clusters.

## Purpose

Verifies that topology placement constraints specified on a workload are correctly propagated to and evaluated by the worker cluster in a MultiKueue setup.

## What's tested

- `topologyRequest` on workloads dispatched through MultiKueue
- Worker cluster topology assignment is reflected in the manager cluster workload status
- TAS constraints are satisfied by the worker cluster's node topology before admission
- Incompatible topology (worker cluster lacks the required topology) causes the workload to be held or dispatched to a different worker
