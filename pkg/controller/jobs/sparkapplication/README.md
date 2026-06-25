# pkg/controller/jobs/sparkapplication/

Integration adapter for `SparkApplication` (`sparkoperator.k8s.io/v1beta2`). Manages Apache Spark jobs submitted via the Spark Operator.

## GenericJob Implementation

| Method | Behavior |
|---|---|
| `IsSuspended()` | Checks if SparkApplication has suspension annotation |
| `Suspend()` | Sets suspension annotation |
| `RunWithPodSetsInfo()` | Removes suspension; injects node affinity into driver + executor templates |
| `PodSets()` | Two PodSets: `driver` (×1) and `executor` (×N) |
| `Finished()` | Returns true when `sparkapp.Status.AppState.State` is `COMPLETED` or `FAILED` |

## Pod Set Mapping

```
SparkApplication.spec:
  driver.coreRequest / memory   → PodSet "driver" (count: 1)
  executor.instances            → PodSet "executor" (count: instances)
```

## MultiKueue Support

Not supported.

## Notes

- Spark driver coordinates; executors perform computation
- Executor count can be dynamic (with Spark autoscaling) — Kueue manages initial quota
