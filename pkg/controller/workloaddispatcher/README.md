# pkg/controller/workloaddispatcher/

WorkloadDispatcher controller. Routes workloads to the appropriate `ClusterQueue` based on dispatch policies, serving as an intermediary layer between job submission and queue assignment.

## Purpose

In large organizations, different teams may submit jobs to a central dispatcher that automatically routes to the correct ClusterQueue based on:
- Namespace of the submitting workload
- Labels on the workload
- ResourceFlavor requirements
- Configured dispatch rules

## Controller Responsibilities

- Watch for workloads that have no `queueName` (or a dispatcher-managed queue)
- Evaluate dispatch rules to select a target `LocalQueue`
- Assign `queueName` to the workload
- Reject workloads that match no dispatch rule

## Feature Gate

`features.WorkloadDispatcher` — disabled by default, opt-in via feature gate.
