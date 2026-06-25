# pkg/visibility/

Implements the Kueue Visibility API server — an aggregated Kubernetes API that provides live queue inspection without requiring etcd storage.

## Architecture

The visibility API is registered as a Kubernetes `APIService`. API requests are routed by the aggregation layer to Kueue's in-process HTTP server:

```
kubectl get pendingworkloads -n team-a
  → kube-apiserver
  → APIService routing
  → Kueue visibility server (in-process)
  → pkg/cache/queue.Manager.PendingWorkloads()
  → Response
```

## Key Type: `VisibilityServer`

Implements `k8s.io/apiserver/pkg/server.GenericAPIServer`. Serves:
- `GET /apis/visibility.kueue.x-k8s.io/v1beta1/clusterqueues/{name}/pendingworkloads`
- `GET /apis/visibility.kueue.x-k8s.io/v1beta2/clusterqueues/{name}/pendingworkloads`
- `GET /apis/visibility.kueue.x-k8s.io/v1beta2/namespaces/{ns}/localqueues/{name}/pendingworkloads`

## Sub-packages

| Package | Purpose |
|---|---|
| [`storage/`](storage/) | REST storage backends implementing the API registry |

## RBAC

Access controlled by standard Kubernetes RBAC:
```yaml
rules:
- apiGroups: ["visibility.kueue.x-k8s.io"]
  resources: ["clusterqueues/pendingworkloads"]
  verbs: ["get"]
```

## Registration

The visibility server is started alongside the controller manager. It registers itself as an `APIService` on startup if the feature is enabled.
