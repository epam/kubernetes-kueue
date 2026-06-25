# test/integration/singlecluster/controller/dra/

Integration tests for Dynamic Resource Allocation (DRA) support in Kueue.

## Purpose

DRA allows workloads to request devices (GPUs, FPGAs, network adapters) via `ResourceClaim` objects rather than extended resources. Kueue must account for device quota and gate workload admission until claims are allocated.

## What's tested

- Workloads with `resourceClaims` in pod specs are admitted only after DRA quota is available
- `ResourceClaim` objects are created by Kueue on behalf of the workload before pods start
- DRA quota is released when the workload finishes
- DRA resources participate in fair sharing (DRS computation includes device requests)
- Feature gate (`DynamicResourceAllocation`) correctly controls DRA behaviour
