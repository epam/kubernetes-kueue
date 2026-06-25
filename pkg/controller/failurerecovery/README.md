# pkg/controller/failurerecovery/

Failure recovery controller. Handles re-admission of workloads that failed due to node failures, pod evictions, or transient infrastructure issues — distinct from user-triggered cancellation.

## Use Case

When a distributed training job fails because a node crashed (not because the job logic failed), it should be automatically re-queued and re-admitted rather than marked permanently failed.

## Controller Responsibilities

- Monitor admitted workloads for failure conditions
- Classify failures: recoverable (infrastructure) vs. terminal (job logic)
- For recoverable failures: reset workload status, re-enqueue for scheduling
- Increment `retryCount` on the workload; stop recovery after `maxRetries`
- Emit events distinguishing recovery from permanent failure

## Key Signals

The controller inspects:
- Pod exit codes and reasons (`OOMKilled`, `Evicted`, `NodeLost` vs. `Error`, `Failed`)
- Node conditions at the time of failure
- Job-type-specific failure signals (via `GenericJob` interface)

## Configuration

Configured per-workload via annotations or per-ClusterQueue via policy objects. The feature is gated by `features.FailureRecovery`.
