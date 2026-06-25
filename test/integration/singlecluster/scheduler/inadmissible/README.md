# test/integration/singlecluster/scheduler/inadmissible/

Integration tests for inadmissible workload handling and requeueing.

## What's tested

- Workloads that cannot be admitted (over quota, missing flavors, failed admission checks) are placed in the `inadmissible` queue
- The backoff-based requeueing controller retries them at increasing intervals
- When the blocking condition is resolved (quota freed, flavor created, check approved), the workload is re-admitted
- `Workload.status.conditions[Evicted]` is set correctly during requeueing cycles
