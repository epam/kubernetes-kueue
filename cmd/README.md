# cmd/

Binary entry points for all Kueue executables.

## Binaries

| Directory | Binary | Purpose |
|---|---|---|
| [`kueue/`](kueue/) | `kueue` | Main controller manager |
| [`kueuectl/`](kueuectl/) | `kueuectl` | kubectl plugin for kueue operations |
| [`kueueviz/`](kueueviz/) | `kueueviz` | Web UI for visualizing queue state |
| [`kueuectl-docs/`](kueuectl-docs/) | `kueuectl-docs` | Documentation generator for kueuectl |
| [`importer/`](importer/) | `importer` | Import existing workloads into Kueue management |
| [`experimental/`](experimental/) | Various | Experimental sidecar tools |

## Build

```bash
make build          # build all binaries
make kueuectl       # build kueuectl only
make run            # run kueue controller manager locally
```

## Installation

The `kueue` binary is packaged as a container image (`gcr.io/k8s-staging-kueue/kueue`). `kueuectl` is distributed as a kubectl plugin via krew.
