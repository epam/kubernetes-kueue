# client-go/

Auto-generated typed Kubernetes clients for Kueue CRDs. This directory is fully generated and should **not be edited manually** — regenerate with `make generate`.

## Contents

```
client-go/
├── applyconfiguration/   # Apply configuration types (server-side apply)
│   ├── kueue/            # Per-CRD apply config structs
│   └── internal/         # Internal marshaling helpers
├── clientset/            # Versioned clientsets
│   └── versioned/        # ClientSet with typed clients per resource
├── informers/            # SharedInformer factories per resource
├── listers/              # Lister interfaces for each resource type
└── typed/                # Low-level typed client operations (Get/List/Create/etc.)
```

## Usage

```go
import (
    kueueclient "sigs.k8s.io/kueue/client-go/clientset/versioned"
    kueueinformers "sigs.k8s.io/kueue/client-go/informers/externalversions"
)

clientset := kueueclient.NewForConfig(restConfig)
factory := kueueinformers.NewSharedInformerFactory(clientset, 0)
wlInformer := factory.Kueue().V1beta1().Workloads()
```

## When to Use

Most Kueue-internal code uses `controller-runtime`'s generic `client.Client` and does not use these generated clients. The generated clients are primarily useful for:
- External tools integrating with Kueue
- `kueuectl` plugin
- Performance testing and bulk operations

## Regeneration

```bash
make generate
```

This runs `code-generator` tools from `hack/tools/code-generator/`.
