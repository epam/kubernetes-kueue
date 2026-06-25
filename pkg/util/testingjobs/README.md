# pkg/util/testingjobs/

Per-framework test job builders. Each sub-package provides a fluent builder for creating test instances of a specific job type.

## Sub-packages

| Package | Builder Type |
|---|---|
| `job/` | `batch/v1 Job` |
| `jobset/` | `JobSet` |
| `pod/` | `Pod` / pod group |
| `deployment/` | `Deployment` |
| `statefulset/` | `StatefulSet` |
| `raycluster/` | `RayCluster` |
| `rayjob/` | `RayJob` |
| `rayservice/` | `RayService` |
| `leaderworkerset/` | `LeaderWorkerSet` |
| `mpijob/` | `MPIJob` |
| `appwrapper/` | `AppWrapper` |
| `sparkapplication/` | `SparkApplication` |
| `trainjob/` | `TrainJob` |
| `pytorchjob/` | `PyTorchJob` |
| `tfjob/` | `TFJob` |
| `jaxjob/` | `JAXJob` |
| `paddlejob/` | `PaddleJob` |
| `xgboostjob/` | `XGBoostJob` |
| `node/` | `Node` (for TAS tests) |

## Usage

```go
job := testingjob.MakeJob("my-job", "default").
    Queue("my-queue").
    Request(corev1.ResourceCPU, "2").
    Parallelism(4).
    Obj()
```

Each builder creates a minimal valid object with sensible test defaults, exposing only the fields relevant to Kueue behavior.
