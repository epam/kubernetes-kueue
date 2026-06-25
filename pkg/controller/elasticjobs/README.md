# pkg/controller/elasticjobs/

Controller support for elastic jobs — workloads that can dynamically adjust their size (number of workers) during execution via `WorkloadSlice` objects.

## Concept

Traditional jobs have a fixed pod count. Elastic jobs can scale up or down while running. Kueue represents this with `WorkloadSlice` — a child of the main `Workload` that represents a change in resource requirements.

## WorkloadSlice

A `WorkloadSlice` tracks a delta request:
- The parent `Workload` holds base resource claims
- Each `WorkloadSlice` adds or removes resource claims
- Kueue admits slices sequentially, ensuring quota is available before scaling

## Controller Responsibilities

- Watch for new `WorkloadSlice` objects
- Admit slices when quota is available
- Update the parent `Workload.status` to reflect current total resources
- Handle slice eviction when quota is reclaimed

## Feature Gate

`features.ElasticJobsViaWorkloadSlices` — controls whether this feature is active.
