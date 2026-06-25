# pkg/controller/jobs/kubeflow/jobs/pytorchjob/

Integration adapter for `PyTorchJob` (`kubeflow.org/v1`). The most widely used KubeFlow training job type for PyTorch distributed training.

## Pod Set Mapping

```
PyTorchJob.spec.pytorchReplicaSpecs:
  Master: { replicas: 1 }   → PodSet "Master" (count: 1)
  Worker: { replicas: N }   → PodSet "Worker" (count: N)
```

## GenericJob

Delegates entirely to `kubeflowjob.KubeflowJob`. No PyTorch-specific overrides needed.

## MultiKueue Support

Full MultiKueue support — dispatches PyTorchJob to worker clusters.

## Common Use Case

```yaml
apiVersion: kubeflow.org/v1
kind: PyTorchJob
metadata:
  labels:
    kueue.x-k8s.io/queue-name: gpu-queue
spec:
  pytorchReplicaSpecs:
    Master:
      replicas: 1
    Worker:
      replicas: 7  # 8 GPUs total (1 master + 7 workers)
```
