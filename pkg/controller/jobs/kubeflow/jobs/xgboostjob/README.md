# pkg/controller/jobs/kubeflow/jobs/xgboostjob/

Integration adapter for `XGBoostJob` (`kubeflow.org/v1`). XGBoost distributed training jobs.

## Pod Set Mapping

```
XGBoostJob.spec.xgboostReplicaSpecs:
  Master: { replicas: 1 }  → PodSet "Master"
  Worker: { replicas: N }  → PodSet "Worker"
```

## MultiKueue Support

Not supported.
