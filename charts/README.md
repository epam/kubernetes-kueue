# charts/

Helm charts for Kueue.

## Contents

| Chart | Description |
|---|---|
| `kueue/` | Main Kueue Helm chart — installs the controller manager, CRDs, RBAC, webhooks, and optional components (cert-manager, Prometheus, KueueViz) |

## Quick Install

```bash
# From OCI registry (recommended for production)
helm install kueue oci://registry.k8s.io/kueue/charts/kueue \
  --version="<version>" \
  --create-namespace --namespace=kueue-system

# From local source
helm install kueue charts/kueue/ \
  --create-namespace --namespace=kueue-system
```

See `charts/kueue/README.md` for the full configuration reference.
