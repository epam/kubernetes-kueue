# charts/kueue/templates/

Helm templates for the Kueue chart. Each sub-directory groups related Kubernetes resources.

## Sub-directories

| Directory | Resources |
|---|---|
| `crd/` | CustomResourceDefinition manifests for all 11 Kueue CRDs (ClusterQueue, LocalQueue, Workload, ResourceFlavor, Cohort, AdmissionCheck, MultiKueueConfig, MultiKueueCluster, ProvisioningRequestConfig, WorkloadPriorityClass, Topology) |
| `manager/` | Controller manager Deployment, ConfigMap (`controller_manager_config.yaml`), PodDisruptionBudget, auth proxy Service |
| `rbac/` | All RBAC: manager ClusterRole + binding, leader-election Role + binding, secrets Role + binding, metrics roles, editor/viewer roles for every CRD, batch admin/user aggregated roles |
| `webhook/` | ValidatingWebhookConfiguration, MutatingWebhookConfiguration manifests and webhook Service |
| `certmanager/` | cert-manager Certificate (webhook TLS, metrics TLS, visibility TLS), Issuer — conditionally rendered when `.Values.enableCertManager=true` |
| `internalcert/` | Self-signed TLS Secret — used when cert-manager is not enabled |
| `visibility/` | APIService registrations for `v1beta1.visibility.kueue.x-k8s.io` and `v1beta2.visibility.kueue.x-k8s.io`, plus RBAC for the visibility server |
| `visibility-apf/` | FlowSchema + PriorityLevelConfiguration for API Priority and Fairness on the Visibility aggregated API |
| `prometheus/` | ServiceMonitor for Prometheus scraping of controller manager metrics |
| `kueueviz/` | KueueViz backend Deployment + Service + Ingress, frontend Deployment + Service + Ingress + ConfigMap, ClusterRole + binding |

## Naming

All resource names use `{{ include "kueue.fullname" . }}` which defaults to `kueue-` + release name, overridable via `.Values.fullnameOverride`.
