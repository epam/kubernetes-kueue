# pkg/scheduler/preemption/common/

Shared utilities used by both classical and fair sharing preemption strategies.

## Contents

- **Candidate collection** — gather all workloads that could be victims (admitted workloads in scope)
- **Victim ordering** — sort candidates by eviction preference (least disruption first)
- **Resource checking** — verify that a set of victims frees enough resources for the preemptor
- **Eviction helpers** — patch workload status to remove `admission` and set `Evicted` condition

## Key Functions

- `IssuePreemptions(ctx, preemptor, victims, cache)` — atomically evict selected victims
- `OrderWorkloads(workloads)` — sort victims to minimize disruption
- `FreeEnoughResources(victims, needed)` — greedy check that victim set satisfies request
