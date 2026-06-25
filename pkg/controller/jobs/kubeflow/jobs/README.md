# pkg/controller/jobs/kubeflow/jobs/

Per-framework KubeFlow job adapters. Each directory contains a thin wrapper around the shared `kubeflowjob` base adapter, providing framework-specific type bindings.

## Contents

| Directory | Job Type | Replica Roles |
|---|---|---|
| `pytorchjob/` | PyTorchJob | Master (×1) + Worker (×N) |
| `tfjob/` | TFJob | Chief (×1) + Worker (×N) + PS (×N) |
| `jaxjob/` | JAXJob | Coordinator (×1) + Worker (×N) |
| `paddlejob/` | PaddleJob | Master (×1) + Worker (×N) |
| `xgboostjob/` | XGBoostJob | Master (×1) + Worker (×N) |

Each adapter registers itself with `jobframework.RegisterIntegration()` in `init()`.
