# Priority Boost Controller

The `priority-boost-controller` is an experimental controller that implements
time-sharing fairness for Kueue workloads by **not** setting
`kueue.x-k8s.io/priority-boost` during an initial admit window, then applying a
**negative** boost after that window while the workload remains admitted. That
lets same–base-priority waiters preempt under **`withinClusterQueue: LowerPriority`**.

## Purpose

In a `ClusterQueue` with a fixed quota, multiple workloads of the same priority
and unknown duration can starve peers. With **LowerPriority**, workloads with the
same effective priority do not preempt each other. During `minAdmitDuration` the
controller leaves the annotation unset, so the admitted workload keeps its base
effective priority. After the window, the controller sets a negative annotation
so same–base-priority pending workloads can preempt, producing round-robin style
fairness at admission time.

**Why not `LowerOrNewerEqualPriority` or `Any`?** Those policies allow preemption
against equal or higher effective priority in ways that do not match this
“defer signal until after a minimum admit time” mechanism. This controller is
documented for **`withinClusterQueue: LowerPriority`**.

This component demonstrates an external policy controller for KEP-7990 (priority
boost signal). It is not part of the main Kueue binary and is deployed separately.

**Requires**: the `PriorityBoost` feature gate enabled in Kueue, and the
`ClusterQueue` must use `withinClusterQueue: LowerPriority`.

## How It Works

```
Workload admitted
      │
      ▼
no annotation               ← same effective priority as peers; no same-priority
  (minAdmitDuration)            preemption under LowerPriority
      │
  [window elapses, still admitted]
      │
      ▼
annotation = -boostValue    ← post-window preemption signal (magnitude = boostValue)
      │
  [workload no longer admitted]
      │
      ▼
annotation cleared
```

`EffectivePriority = basePriority + boost`. With no annotation, boost is 0.
After the window, boost is `-boostValue`, lowering effective priority so peers
can preempt.

## Build

```bash
make build
```

```bash
make test
```

```bash
make image-build
```

## Deploy

```bash
kubectl apply -k config
```

## Configuration

The controller reads `--config` YAML:

| Field | Type | Default (in code) | Description |
|-------|------|-------------------|-------------|
| `minAdmitDuration` | duration | `"0"` | How long after admission to leave **no** annotation. After that, while still admitted, set **negative** boost. `"0"` disables the controller behavior. |
| `boostValue` | int32 | `100000` | Magnitude of the **negative** boost after the window (stored as `-boostValue` in the annotation). |
| `workloadSelector` | `LabelSelector` | omit | If set, only workloads whose labels match are managed; others have any boost annotation **removed**. |
| `maxWorkloadPriority` | int32 | omit | If set, workloads with `spec.priority` **greater** than this are out of scope (unset `spec.priority` is treated as `0`). |

Example manifests: `config/manager/controller_config_map.yaml`.

### Example

```yaml
minAdmitDuration: "30m"
boostValue: 100000
# workloadSelector:
#   matchLabels:
#     tier: batch
# maxWorkloadPriority: 1000
```

## Helm

Chart path: `charts/priority-boost-controller`
