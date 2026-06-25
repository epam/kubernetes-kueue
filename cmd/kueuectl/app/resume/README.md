# cmd/kueuectl/app/resume/

`kueuectl resume` subcommands for resuming stopped queues.

## Commands

### `kueuectl resume clusterqueue <name>`

Sets `spec.stopPolicy=None` — resumes accepting and admitting workloads.

### `kueuectl resume localqueue <name>`

Sets `spec.stopPolicy=None` on the LocalQueue.
