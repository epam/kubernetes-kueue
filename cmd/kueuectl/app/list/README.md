# cmd/kueuectl/app/list/

`kueuectl list` subcommands for listing Kueue resources.

## Commands

### `kueuectl list workloads`

Lists `Workload` objects with Kueue-specific formatting:
- Status column (Pending/Admitted/Finished)
- Queue name
- Priority
- Wait time (time since creation)
- Admitted time

### `kueuectl list localqueues`

Lists `LocalQueue` objects with queue statistics:
- Status (Active/Inactive)
- ClusterQueue reference
- Pending/Admitted workload counts

### `kueuectl list clusterqueues`

Lists `ClusterQueue` objects with resource usage summary:
- Status
- Cohort
- Resource usage vs. quota

## Table Printing

Uses `k8s.io/cli-runtime/pkg/printers` for tabular output consistent with `kubectl get` formatting.
