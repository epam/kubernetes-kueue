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
- `config.minAdmitDuration` — duration with no annotation after admission; then negative boost while admitted
- `config.boostValue` — magnitude of the negative `kueue.x-k8s.io/priority-boost` after the window
- `config.workloadSelector` — optional `LabelSelector`; workloads that do not match are not managed (annotation cleared)
- `config.maxWorkloadPriority` — optional; workloads with higher `spec.priority` are excluded (omit key for no limit)
- `kueue.enabled` — optional bundled Kueue subchart
