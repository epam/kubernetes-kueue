# pkg/util/kubeversion/

Kubernetes API version detection utilities.

## Key Functions

- `DetectedVersion(ctx, discoveryClient) (*version.Version, error)` — discover the running Kubernetes server version
- `SupportsAPIGroup(ctx, discoveryClient, group, version string) (bool, error)` — check if an API group/version is registered in the cluster

## Usage

Job integration adapters use this to check whether the target API (e.g., `ray.io/v1`) is available in the cluster before registering their controllers:

```go
if ok, err := kubeversion.SupportsAPIGroup(ctx, disco, "ray.io", "v1"); !ok {
    // Ray CRDs not installed, skip Ray integration
}
```
