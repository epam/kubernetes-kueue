# Priority Boost Controller

The `priority-boost-controller` is an experimental controller that implements
time-sharing fairness for Kueue Workloads by temporarily boosting the effective
priority of admitted workloads, preventing them from being preempted by
same-priority peers for a configurable minimum duration.

## Purpose

In a ClusterQueue with a fixed quota, multiple workloads of the same priority
and unknown execution duration can lead to starvation: the first admitted
workload monopolizes resources while others wait indefinitely.

This controller addresses that by:

1. Setting `kueue.x-k8s.io/priority-boost` to a large value (`boostValue`)
   the moment a workload is admitted, making it appear higher-priority than
   any same-base-priority pending workload.
2. Keeping the boost active for `minAdmitDuration` (e.g., 30 minutes),
   guaranteeing the workload cannot be preempted by same-priority peers during
   that window.
3. Removing the boost once the window expires, restoring normal preemption
   eligibility — enabling round-robin style fairness at the admission level.

This component demonstrates an external policy controller for KEP-7990
(priority boost signal). It is not part of the main Kueue binary and is
intended to be built and deployed independently.

**Requires**: the `PriorityBoost` feature gate must be enabled in Kueue, and
the ClusterQueue must have `withinClusterQueue: LowerOrNewerEqualPriority` (or
`Any`) preemption configured.

## How It Works

```
Workload admitted
      │
      ▼
boost = boostValue          ← controller sets priority-boost annotation
      │
  [minAdmitDuration elapses]
      │
      ▼
boost removed               ← controller removes annotation; workload becomes
                              preemptable by same-priority peers
```

Because `EffectivePriority = basePriority + boost`, a pending workload with
`basePriority = 0` cannot satisfy `LowerOrNewerEqualPriority` against an
admitted workload with `basePriority = 0, boost = 100000` — so no preemption
occurs until the window expires.

## Build

To build the `priority-boost-controller` binary:

```bash
make build
```

To run controller tests:

```bash
make test
```

To build the container image:

```bash
make image-build
```

## Deploy

Apply manifests from `config/`:

```bash
kubectl apply -k config
```

## Configuration

The controller reads configuration from the `--config` YAML file:

| Field              | Type     | Default   | Description |
|--------------------|----------|-----------|-------------|
| `minAdmitDuration` | duration | `"30m"`   | Minimum time a workload must be admitted before it can be preempted by same-priority peers. Set to `"0"` to disable. |
| `boostValue`       | int32    | `100000`  | Priority boost applied during the protection window. Must exceed the highest base priority in the queue. |

Default config is defined in `config/manager/controller_config_map.yaml`.

### Example

```yaml
minAdmitDuration: "30m"
boostValue: 100000
```

## Helm

A Helm chart is available at:

`charts/priority-boost-controller`
