# pkg/resources/

Resource quantity math utilities for computing and comparing Kubernetes resource amounts.

## Key Types

### `FlavorResourceQuantities`

```go
type FlavorResourceQuantities map[ResourceFlavorReference]map[corev1.ResourceName]resource.Quantity
```

Maps `(flavor, resource)` → quantity. Used throughout the scheduler and cache to track resource usage and quotas.

### `FlavorResourceSet`

Set variant for presence checks without quantities.

## Key Functions

- `Add(a, b FlavorResourceQuantities) FlavorResourceQuantities` — sum resource maps
- `Sub(a, b FlavorResourceQuantities)` — subtract (for usage calculation)
- `Fits(usage, quota FlavorResourceQuantities) bool` — check if usage is within quota
- `DominantShare(usage, total FlavorResourceQuantities) float64` — compute DRS value

## Usage

Resource tracking is ubiquitous in Kueue:
```go
// Check if adding a workload would exceed quota
if !resources.Fits(cq.usage + wl.TotalRequests, cq.quota) {
    return NoFit
}
```
