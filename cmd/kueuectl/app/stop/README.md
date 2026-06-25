# cmd/kueuectl/app/stop/

`kueuectl stop` subcommands for stopping ClusterQueues and LocalQueues.

## Commands

### `kueuectl stop clusterqueue <name>`

Sets `spec.stopPolicy=Hold` (or `HoldAndDrain` with `--drain`). Prevents new workloads from being admitted. Existing admitted workloads continue running unless `--drain` is specified.

### `kueuectl stop localqueue <name>`

Sets `spec.stopPolicy=Hold` on the LocalQueue. Blocks new submissions to this queue.

## Options

- `--drain` — also evict currently admitted workloads (`HoldAndDrain`)
