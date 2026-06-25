# test/integration/singlecluster/controller/failurerecovery/

Integration tests for pod failure recovery.

## Purpose

When pods in an admitted workload fail (node crash, OOM kill), Kueue can optionally re-admit the workload rather than letting it stay in a broken state. This is controlled by the `FailureRecovery` feature gate.

## What's tested

- Pod termination events trigger the failure recovery controller
- Failed workloads are evicted and re-queued automatically
- `ObjectRetentionPolicies` control how long failed workloads are retained before recovery
- Re-admission uses the same quota as the original admission (no double-counting)
- The feature can be disabled — workloads stay in their failed state when the gate is off
