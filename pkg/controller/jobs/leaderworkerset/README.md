# pkg/controller/jobs/leaderworkerset/

Integration adapter for `LeaderWorkerSet` (`leaderworkerset.sigs.k8s.io/v1`). A `LeaderWorkerSet` creates N identical groups, each with one leader pod and W worker pods — designed for LLM training and inference.

## GenericJob Implementation

| Method | Behavior |
|---|---|
| `IsSuspended()` | Checks `lws.Spec.LeaderWorkerTemplate.RestartPolicy == "Never"` or suspension annotation |
| `Suspend()` | Sets suspension annotation |
| `RunWithPodSetsInfo()` | Removes suspension; injects node affinity into leader + worker templates |
| `PodSets()` | Two PodSets: leader template × 1 per group, worker template × W per group |

## Pod Group Structure

```
LeaderWorkerSet (replicas: N, size: W+1):
  group-0: leader-pod-0 + worker-pod-0-{0..W-1}
  group-1: leader-pod-1 + worker-pod-1-{0..W-1}
  ...
  group-N-1: leader-pod-(N-1) + worker-pod-(N-1)-{0..W-1}
```

Kueue accounts for `N × (1 + W)` pods total.

## MultiKueue Support

Not supported.

## TAS Integration

LeaderWorkerSet works with TAS to ensure each group's pods land on nodes within the same topology domain (e.g., same rack for high-bandwidth communication).
