# cmd/kueuectl/

kubectl plugin for Kueue operations. Extends `kubectl` with Kueue-specific commands.

## Installation

```bash
# Via krew:
kubectl krew install kueue

# Or build directly:
make kueuectl
```

## Commands

```
kueuectl create localqueue    # Create a LocalQueue
kueuectl create clusterqueue  # Create a ClusterQueue
kueuectl list workloads       # List workloads across queues
kueuectl list localqueues     # List local queues
kueuectl list clusterqueues   # List cluster queues
kueuectl stop clusterqueue    # Stop a ClusterQueue (drain)
kueuectl stop localqueue      # Stop a LocalQueue
kueuectl resume clusterqueue  # Resume a stopped ClusterQueue
kueuectl resume localqueue    # Resume a stopped LocalQueue
kueuectl delete workload      # Delete a workload
kueuectl pass-through         # Pass-through to kubectl for kueue resources
kueuectl version              # Show kueuectl version
```

## Entry Point

`main.go` creates the root cobra command and delegates to `app/` for command implementations.

## Sub-packages

| Package | Purpose |
|---|---|
| [`app/`](app/) | Root command and subcommand implementations |
