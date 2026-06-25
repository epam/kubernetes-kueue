# cmd/kueuectl/app/passthrough/

Pass-through commands to `kubectl` for Kueue resources. Enables familiar `kubectl`-style operations on Kueue CRDs.

## Purpose

For resources where `kueuectl` doesn't have specialized subcommands, pass-through forwards to `kubectl` with the correct API group and resource name:

```bash
kueuectl pass-through get workloads     # → kubectl get workloads.kueue.x-k8s.io
kueuectl pass-through describe cq my-cq # → kubectl describe clusterqueues my-cq
```

## Supported Resources

All Kueue CRDs are aliased for pass-through: `workloads`, `clusterqueues`, `localqueues`, `resourceflavors`, `admissionchecks`, etc.
