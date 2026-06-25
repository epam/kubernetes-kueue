# pkg/controller/jobs/kubeflow/jobs/paddlejob/

Integration adapter for `PaddleJob` (`kubeflow.org/v1`). PaddlePaddle distributed training jobs.

## Pod Set Mapping

```
PaddleJob.spec.paddleReplicaSpecs:
  Master: { replicas: 1 }  → PodSet "Master"
  Worker: { replicas: N }  → PodSet "Worker"
```

## MultiKueue Support

Not supported.
