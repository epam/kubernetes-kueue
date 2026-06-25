# test/e2e/dra/

End-to-end tests for Dynamic Resource Allocation (DRA) integration with Kueue.

## Purpose

Verifies that Kueue correctly accounts for and gates DRA `ResourceClaim` resources attached to workload pods. DRA is a Kubernetes feature that allows heterogeneous device resources (GPUs, FPGAs, network devices) to be requested via `ResourceClaim` objects rather than extended resources.

## Sub-packages

### `baseline/`

Core DRA scenarios:
- Workloads with `resourceClaims` in pod specs are admitted only when device quota is available
- `ResourceClaim` objects are created and bound by the DRA driver before the workload runs
- Preemption correctly releases device claims

## Prerequisites

A DRA driver must be installed. Tests use a fake DRA driver (`test/dra-driver/`) that simulates device allocation without real hardware.
