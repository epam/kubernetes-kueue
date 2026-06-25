# test/integration/singlecluster/importer/

Integration tests for the Kueue importer tool.

## Purpose

Verifies that the importer correctly ingests pre-existing Pods into Kueue — creating Workloads, setting admission status, and adding Kueue labels — without needing to run the full importer binary.

## What's tested

- Pod selection by label
- Mapping rules (label-based and file-based)
- Workload creation with pre-admitted status
- `--dry-run` mode (check but don't import)
- Duplicate-import prevention (idempotency)
- Missing ClusterQueue / LocalQueue detection during check phase
