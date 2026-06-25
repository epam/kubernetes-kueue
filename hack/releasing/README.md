# hack/releasing/

Release automation scripts for Kueue releases.

## Release Workflow

Kueue follows the Kubernetes release process. Releases are staged to a GCS bucket and then promoted to the final registry.

## Scripts

| Script | Purpose |
|---|---|
| `prepare_pull.sh` | Prepare a release PR: bump version strings, update CHANGELOG, generate manifests |
| `ci_pull.sh` | CI script that runs release validation on a release PR |
| `promote_pull.sh` | Promote a staged release to the final registry (runs in CI after approval) |
| `wait_for_images.sh` | Poll GCS until release images are available (used between stage and promote) |
| `generate_krew.sh` | Generate the `krew` plugin manifest for kueuectl after a release |
| `sync-notes.sh` | Sync release notes from GitHub to the Kueue website |
