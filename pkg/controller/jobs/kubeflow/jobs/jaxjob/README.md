# pkg/controller/jobs/kubeflow/jobs/jaxjob/

Integration adapter for `JAXJob` (`kubeflow.org/v1`). JAX distributed training jobs.

## Pod Set Mapping

```
JAXJob.spec.jaxReplicaSpecs:
  Coordinator: { replicas: 1 }  → PodSet "Coordinator"
  Worker:      { replicas: N }  → PodSet "Worker"
```

## Notes

JAX training uses a coordinator-worker pattern. The coordinator manages distributed state; workers run computation. Kueue accounts for all replicas.

## MultiKueue Support

Not supported.
