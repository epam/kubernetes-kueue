# pkg/controller/jobs/kubeflow/jobs/tfjob/

Integration adapter for `TFJob` (`kubeflow.org/v1`). TensorFlow distributed training.

## Pod Set Mapping

```
TFJob.spec.tfReplicaSpecs:
  Chief:  { replicas: 1 }   → PodSet "Chief"
  PS:     { replicas: N }   → PodSet "PS" (parameter servers)
  Worker: { replicas: M }   → PodSet "Worker"
```

## Notes

TFJob supports three-role distributed training (Chief, PS, Worker). The `PS` (Parameter Server) role is optional for parameter server strategy; for all-reduce strategy, only Chief + Worker are used.

## MultiKueue Support

Full MultiKueue support.
