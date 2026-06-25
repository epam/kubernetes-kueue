# pkg/util/taints/

Node taint and toleration utilities for Kueue's node flavor matching.

## Key Functions

- `ToleratesAll(tolerations []corev1.Toleration, taints []corev1.Taint) bool` — check if tolerations satisfy all taints
- `FilterTolerations(tolerations, existingTolerations) []corev1.Toleration` — return only tolerations not already present
- `TaintToToleration(taint corev1.Taint) corev1.Toleration` — convert a taint to a matching toleration

## Usage

When `ResourceFlavor` has `nodeTaints`, the flavor assigner calls this package to determine whether a workload's pod spec already tolerates those taints, and if not, injects the required tolerations.
