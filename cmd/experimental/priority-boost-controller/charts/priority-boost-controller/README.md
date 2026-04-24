# priority-boost-controller

Helm chart for deploying the experimental priority-boost controller.

## Install

```bash
helm install priority-boost-controller ./charts/priority-boost-controller \
  --namespace kueue-system \
  --create-namespace
```

## Values

Main values are under `priorityBoostController`:

- `image.repository`, `image.tag`, `image.pullPolicy`, `imagePullSecrets`
- `config.minAdmitDuration` — after a workload is **admitted**, the controller waits this long before
  touching `kueue.x-k8s.io/priority-boost`. During the wait it **requeues** with `RequeueAfter`
  (roughly the remaining time); it does **not** rewrite the annotation on a timer. Once the window
  has passed while the workload stays admitted, the annotation is set **once** to `-boostValue` and
  left stable (no periodic updates) until admission changes, the workload falls out of selector /
  max priority, or the controller clears it.
- `config.boostValue` — unsigned magnitude; after the window Kueue sees annotation `-boostValue`.
  Effective priority is `spec.priority` plus the annotation integer
  (`pkg/util/priority.ParseEffectivePriority`). Example: `spec.priority` = `1000`, `boostValue` =
  `100000` → annotation `-100000` → effective priority `1000 + (-100000) = -99000`, which makes the
  running workload easier to preempt when within-cluster preemption uses lower effective priority.
  The annotation is negative so the workload becomes *less* urgent; the Helm key stays positive as
  the magnitude used to lower effective priority.
- `config.workloadSelector` — optional `LabelSelector`; workloads that do not match are not
  managed (annotation cleared)
- `config.maxWorkloadPriority` — optional; workloads with higher `spec.priority` are excluded (omit key for no limit)
- `kueue.enabled` — optional bundled Kueue subchart
