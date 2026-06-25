# pkg/controller/jobs/

Job integration adapters. Each subdirectory implements the `jobframework.GenericJob` interface for a specific Kubernetes job type, enabling Kueue to manage that job type's lifecycle (suspend, resume, quota, preemption).

## Supported Job Types

| Directory | Job Type | API Group | MultiKueue |
|---|---|---|---|
| `job/` | `batch/v1 Job` | `batch/v1` | Yes |
| `jobset/` | `JobSet` | `jobset.x-k8s.io/v1alpha2` | Yes |
| `pod/` | Plain `Pod` / Pod groups | `core/v1` | No |
| `deployment/` | `Deployment` | `apps/v1` | No |
| `statefulset/` | `StatefulSet` | `apps/v1` | No |
| `raycluster/` | `RayCluster` | `ray.io/v1` | Yes |
| `rayjob/` | `RayJob` | `ray.io/v1` | Yes |
| `rayservice/` | `RayService` | `ray.io/v1` | No |
| `leaderworkerset/` | `LeaderWorkerSet` | `leaderworkerset.sigs.k8s.io/v1` | No |
| `trainjob/` | `TrainJob` | `trainer.kubeflow.org/v1alpha1` | Yes |
| `appwrapper/` | `AppWrapper` | `workload.codeflare.dev/v1beta2` | No |
| `sparkapplication/` | `SparkApplication` | `sparkoperator.k8s.io/v1beta2` | No |
| `mpijob/` | `MPIJob` | `kubeflow.org/v2beta1` | No |
| `kubeflow/` | All KubeFlow training operators | `kubeflow.org` | Yes (some) |

## Common Pattern

Every adapter:
1. Defines a struct embedding the job object: `type Job struct { v1.Job }`
2. Implements all `GenericJob` methods
3. Registers via `jobframework.RegisterIntegration()` in `init()`
4. Sets up a validating/mutating webhook
5. Sets up field indexes for fast workload→job lookup

## Enabling Integrations

Each framework must be listed in `Configuration.integrations.frameworks`:
```yaml
integrations:
  frameworks:
  - "batch/v1"
  - "ray.io/v1"
```

Frameworks not listed are ignored even if the CRD exists in the cluster.
