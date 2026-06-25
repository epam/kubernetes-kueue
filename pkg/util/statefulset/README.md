# pkg/util/statefulset/

StatefulSet utility helpers.

## Key Functions

- `GetPodTemplate(sts *appsv1.StatefulSet) *corev1.PodTemplateSpec` — return the pod template
- `GetReplicas(sts *appsv1.StatefulSet) int32` — return replica count (defaulting nil to 1)
- `SetReplicas(sts *appsv1.StatefulSet, n int32)` — set replica count

## Usage

Used by the StatefulSet adapter (`pkg/controller/jobs/statefulset/`) for consistent access to StatefulSet fields.
