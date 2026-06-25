# pkg/util/resource/

Resource quantity helpers for working with `resource.Quantity` values.

## Key Functions

- `Sum(quantities ...resource.Quantity) resource.Quantity` — sum multiple quantities
- `IsZero(q resource.Quantity) bool` — check if quantity is zero
- `Cmp(a, b resource.Quantity) int` — compare quantities
- `MergeResourceList(a, b corev1.ResourceList) corev1.ResourceList` — merge resource lists (b overrides a)
- `RequestsForPod(pod corev1.Pod) corev1.ResourceList` — sum all container resource requests in a pod

## Usage

```go
total := resource.Sum(container1.Requests["cpu"], container2.Requests["cpu"])
if resource.IsZero(podRequests["memory"]) {
    // pod has no memory request — LimitRange defaults apply
}
```
