# pkg/webhooks/

Mutating and validating admission webhooks for all Kueue CRDs. Webhooks run inside the controller manager process and intercept API server requests before objects are stored.

## Webhooks Defined

| Webhook | Resource | Type | Purpose |
|---|---|---|---|
| `ClusterQueueWebhook` | ClusterQueue | Mutating + Validating | Default fields; validate quota/preemption config |
| `LocalQueueWebhook` | LocalQueue | Mutating + Validating | Default namespace; validate CQ reference |
| `WorkloadWebhook` | Workload | Mutating + Validating | Immutable field checks; priority injection |
| `ResourceFlavorWebhook` | ResourceFlavor | Validating | Validate node labels/taints |
| `AdmissionCheckWebhook` | AdmissionCheck | Validating | Validate controller name |
| `MultiKueueClusterWebhook` | MultiKueueCluster | Mutating + Validating | Validate secret reference |
| `CohortWebhook` | Cohort | Validating | Validate hierarchy |
| `TopologyWebhook` | Topology | Validating | Validate level structure |

## Common Validations

- **Immutability**: Fields like `spec.clusterQueue` on LocalQueue cannot change while workloads are admitted
- **Reference validity**: ClusterQueue references in LocalQueue must match existing CQs
- **Resource format**: Resource requests must be valid Kubernetes `resource.Quantity` values
- **Field constraints**: `resourceGroups` max 16 entries, max 256 total flavors across all groups

## TLS

Webhooks are served over TLS. Certificate management options:
1. `InternalCertManagement` — Kueue self-signs and rotates its own TLS certificates
2. cert-manager — external certificate manager handles TLS (disable internal cert management)

## Registration

All webhooks are registered in `cmd/kueue/main.go` via `webhooks.Setup(mgr)`. They run on the webhook server port (default: `:9443`).
