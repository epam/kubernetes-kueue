# test/integration/singlecluster/tas/

Integration tests for Topology-Aware Scheduling (TAS).

## Purpose

Verifies that the TAS controller correctly enforces topology placement constraints on workloads. Unlike E2E tests, these run in envtest without real nodes — topology awareness is simulated by creating Node objects with topology labels.

## What's tested

- `Topology` CRD creation and validation
- `ResourceFlavor.topologyName` wiring to a Topology object
- Workload admission with `topologyRequest` (Required / BestEffort)
- Topology assignment stored in `Workload.Status.admission.podSetAssignments[].topologyAssignment`
- Workload is not admitted when topology constraints cannot be satisfied
- TopologyUngater controller un-gates pods when topology conditions are met
- Hot-swap: reassigning topology assignment to a different set of nodes
- TAS with multiple PodSets (e.g., leader + workers)
