# apis/config/

Controller manager configuration API. These types are loaded at startup from a `Configuration` object (not stored as CRDs in etcd but loaded from a ConfigMap or file).

## Versions

| Version | Status | Notes |
|---|---|---|
| `v1beta1` | Deprecated | Supported for backwards compatibility |
| `v1beta2` | Current | Use for all new installations |

## Key Type: `Configuration`

Defined in `v1beta2/configuration_types.go`. Fields include:

- **`Namespace`** — namespace where Kueue runs (default: `kueue-system`)
- **`ControllerManager`** — controller-runtime manager settings (leader election, metrics port, health probe port)
- **`InternalCertManagement`** — whether Kueue manages its own TLS certs for webhooks
- **`ClientConnection`** — burst/QPS for the Kubernetes API client
- **`Integrations`** — which job frameworks to enable (e.g., `batch/v1`, `ray.io/v1`)
- **`QueueVisibility`** — depth of pending workload visibility per queue
- **`MultiKueue`** — MultiKueue-specific settings (worker namespace, GC interval)
- **`FairSharing`** — fair sharing strategy (`None`, `Preempt`)
- **`Resources`** — resource transformer configuration

## Usage

```yaml
apiVersion: config.kueue.x-k8s.io/v1beta2
kind: Configuration
namespace: kueue-system
integrations:
  frameworks:
  - "batch/v1"
  - "jobset.x-k8s.io/v1alpha2"
```

The configuration is loaded by `pkg/config` and applied to the manager before controllers start.
