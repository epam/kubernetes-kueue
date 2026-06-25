# pkg/controller/jobs/kubeflow/

KubeFlow Training Operator integrations. Contains adapters for PyTorchJob, TFJob, JAXJob, PaddleJob, XGBoostJob, and MPIJob.

## Structure

```
kubeflow/
├── jobs/                    # Per-framework adapters
│   ├── jaxjob/
│   ├── paddlejob/
│   ├── pytorchjob/
│   ├── tfjob/
│   └── xgboostjob/
└── kubeflowjob/             # Shared base adapter for all KubeFlow job types
```

## Common Pattern

All KubeFlow training jobs follow the same pattern:
- `spec.{Master|Chief|Launcher}ReplicaSpec` — coordinator role (count: 1)
- `spec.{Worker}ReplicaSpec` — worker roles (count: N)
- `spec.runPolicy.suspend` — suspension control

All adapters delegate to `kubeflowjob/` for the common implementation.
