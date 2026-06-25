# pkg/controller/jobs/ray/

Shared utilities used by all Ray job adapters (`raycluster/`, `rayjob/`, `rayservice/`).

## Contents

- Common Ray-specific constants and label keys
- Shared helper functions for extracting PodSets from Ray cluster specs
- Shared webhook validation logic for Ray resources
- Node affinity injection helpers for head + worker group templates

## Why Separate

Ray's three resource types (`RayCluster`, `RayJob`, `RayService`) share significant structural overlap in how they define pod templates. This package prevents duplication across the three adapters.
