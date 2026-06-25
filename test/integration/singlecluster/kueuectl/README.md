# test/integration/singlecluster/kueuectl/

Integration tests for kueuectl commands against a live API server (envtest).

## Purpose

Unlike unit tests in `cmd/kueuectl/app/*/` which use fake clients, these tests run kueuectl commands against a real API server started by envtest. This verifies that command output and API interactions work correctly with actual serialisation, watches, and status subresources.

## What's tested

- `kueuectl list clusterqueue` — output format, filtering, sorting
- `kueuectl list localqueue` — per-namespace listing
- `kueuectl list workload` — status, admitted/pending filtering
- `kueuectl create localqueue` / `clusterqueue` / `resourceflavor`
- `kueuectl stop workload` / `resume workload`
- `kueuectl delete workload`
- `kueuectl pass-through` — generic kubectl-style passthrough for Kueue resources

## Test helpers

`util.go` — shared helpers for running kueuectl commands with a test kubeconfig, capturing stdout/stderr, and asserting on output.
