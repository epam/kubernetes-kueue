# pkg/util/tolerations/

Additional toleration utilities beyond what `pkg/util/taints` provides.

## Key Functions

- `MergeTolerations(a, b []corev1.Toleration) []corev1.Toleration` — merge toleration lists, deduplicating
- `HasToleration(tolerations []corev1.Toleration, key, value string) bool` — check for a specific toleration

## Usage

Used when injecting ResourceFlavor tolerations into job pod templates during the `RunWithPodSetsInfo` phase.
