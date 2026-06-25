# pkg/controller/jobs/appwrapper/

Integration adapter for `AppWrapper` (`workload.codeflare.dev/v1beta2`). AppWrapper is a CodeFlare framework that wraps multiple Kubernetes resources into a single schedulable unit.

## GenericJob Implementation

| Method | Behavior |
|---|---|
| `IsSuspended()` | Checks `appwrapper.Spec.Suspend` |
| `Suspend()` | Sets `Spec.Suspend = true` |
| `RunWithPodSetsInfo()` | Unsuspends; injects node affinity into each component's pod template |
| `PodSets()` | One PodSet per component in the AppWrapper |

## AppWrapper Components

An AppWrapper contains an arbitrary list of Kubernetes resources (Jobs, Services, ConfigMaps). Each component that has pods contributes to the PodSet count.

```yaml
spec:
  components:
  - template:
      apiVersion: batch/v1
      kind: Job
      spec: ...  # contributes pods
```

## MultiKueue Support

Not supported.

## Use Case

AppWrapper is commonly used in CodeFlare/MCAD stacks where multiple resources (Jobs, Services, ConfigMaps) must be co-scheduled together as a unit.
