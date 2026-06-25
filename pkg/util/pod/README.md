# pkg/util/pod/

Pod utility functions.

## Key Functions

- `TotalRequests(pod corev1.Pod) corev1.ResourceList` — sum all container requests including init containers (taking max, not sum)
- `FindContainer(pod, name) (*corev1.Container, bool)` — find a named container
- `IsTerminated(pod) bool` — returns true if pod is in `Succeeded` or `Failed` phase
- `HasSchedulingGate(pod, gate string) bool` — check for a specific scheduling gate
- `RemoveSchedulingGate(pod, gate string)` — remove a gate from pod spec

## Init Container Handling

Init containers run sequentially but are counted toward the peak resource usage (take max of init containers, then add to regular containers). This matches how the Kubernetes scheduler accounts for init containers.
