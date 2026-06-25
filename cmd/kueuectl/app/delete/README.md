# cmd/kueuectl/app/delete/

`kueuectl delete` subcommands.

## Commands

### `kueuectl delete workload <name>`

Deletes a `Workload` object. If the workload is admitted, this triggers eviction of the underlying job. Equivalent to `kubectl delete workload <name>` but with Kueue-aware messaging.

## Options

- `--namespace` / `-n` — namespace of the workload
- `--wait` — wait for deletion to complete
