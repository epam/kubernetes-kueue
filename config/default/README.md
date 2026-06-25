# config/default/

The standard Kueue installation overlay. Composes all default components into a single deployable configuration.

## What it installs

- All 11 CRDs (from `components/crd/`)
- Controller manager Deployment + RBAC (from `components/manager/` + `components/rbac/`)
- Webhook service + webhook configurations (from `components/webhook/`)
- Internal self-signed TLS secret (from `components/internalcert/`)
- Visibility API service registration (from `components/visibility/`)
- Metrics service

## Patches

| Patch file | What it patches |
|---|---|
| `manager_config_patch.yaml` | Merges user configuration into `controller_manager_config.yaml` |
| `manager_webhook_patch.yaml` | Adds webhook cert volume mounts to the manager pod |
| `manager_metrics_patch.yaml` | Configures the metrics endpoint (TLS, port) |
| `manager_visibility_patch.yaml` | Configures the visibility API service endpoint |
| `mutating_webhookcainjection_patch.yaml` | Injects self-signed CA into the MutatingWebhookConfiguration |
| `validating_webhookcainjection_patch.yaml` | Injects self-signed CA into the ValidatingWebhookConfiguration |
| `apiservice_cainjection_patch.yaml` | Injects self-signed CA into the APIService objects |

## Install

```bash
kubectl apply -k config/default
```
