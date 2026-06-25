# pkg/util/limitrange/

LimitRange constraint checking utilities. Ensures workload resource requests satisfy namespace-level `LimitRange` defaults and limits before admission.

## Key Functions

- `Summarize(lrs []corev1.LimitRange) LimitRangeSummary` — compute effective limits from all LimitRanges in a namespace
- `TotalRequestsWithDefaults(summary, pods) FlavorResourceQuantities` — apply LimitRange defaults to pods without explicit requests
- `Satisfies(requests, summary) error` — validate that requests are within LimitRange bounds

## Why This Matters

If a pod has no CPU/memory request, LimitRange defaults are applied by the admission controller. Kueue must account for these defaults when computing workload resource requirements — otherwise it might admit a workload that actually needs more resources than declared.

## Integration

Called during workload creation and validation in `pkg/controller/jobframework/` to compute accurate resource requests.
