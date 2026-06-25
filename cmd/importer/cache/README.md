# cmd/importer/cache/

Local state cache for the Kueue importer tool. Tracks which resources have already been imported to prevent duplicates.

## Purpose

The importer may be run multiple times (e.g., in a loop with `--continuous` mode or restarted after failure). This cache records which Pods/Jobs have already had their Workloads created, allowing the importer to skip them on subsequent runs.

## Implementation

Uses an in-memory map keyed by Pod UID, backed by a persistent file for crash recovery. On startup, the cache is loaded from the backing file.
