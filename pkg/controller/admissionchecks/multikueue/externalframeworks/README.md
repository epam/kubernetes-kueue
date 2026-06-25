# pkg/controller/admissionchecks/multikueue/externalframeworks/

Plugin mechanism allowing custom job types to integrate with MultiKueue without modifying Kueue itself.

## Purpose

Built-in job adapters (batch/Job, RayJob, etc.) are compiled into Kueue. External frameworks — custom CRD-based job types — cannot be compiled in. This package provides a mechanism to register MultiKueue adapters for arbitrary CRDs at runtime.

## How It Works

1. An external CRD registers itself as a MultiKueue-compatible framework by adding annotations to its `CustomResourceDefinition` object
2. The `externalframeworks` package discovers these annotations at startup
3. It creates a generic `MultiKueueAdapter` for the CRD based on the annotation metadata
4. The generic adapter uses `unstructured.Unstructured` to copy/sync objects without type-specific code

## Configuration

In `Configuration.integrations.externalFrameworks`:
```yaml
integrations:
  externalFrameworks:
  - group: my-operator.io
    version: v1
    kind: MyJob
```

## Limitations

External frameworks use a generic copy strategy — they cannot perform job-type-specific transformations (e.g., injecting node affinities into specific fields). For full control, compile a custom adapter into a Kueue fork.
