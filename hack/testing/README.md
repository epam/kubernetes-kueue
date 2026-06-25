# hack/testing/

Test runner and CI scripts for Kueue.

## Scripts

| Script | Purpose |
|---|---|
| `e2e-test.sh` | Run single-cluster E2E tests against a kind cluster |
| `e2e-common.sh` | Shared functions for E2E cluster setup (image loading, Kueue deploy, cleanup) |
| `e2e-common_test.sh` | Unit tests for e2e-common.sh functions |
| `e2e-multikueue-test.sh` | Run MultiKueue E2E tests (creates 3 kind clusters) |
| `e2e-kueueviz-backend.sh` | Run KueueViz backend E2E tests |
| `e2e-kueueviz-frontend.sh` | Run KueueViz frontend E2E tests |
| `e2e-kueueviz-local.sh` | Start KueueViz locally against a kind cluster for manual testing |
| `performance-test.sh` | Run scheduler performance benchmarks |
| `compare-performance.sh` | Compare two performance benchmark result CSVs |
| `get-build-logs.sh` | Fetch Prow CI build logs for a PR |
| `retry.sh` | Generic retry wrapper with exponential backoff |
| `shard-integration-tests.sh` | Split integration test packages across CI shards |

## Sub-directories

| Directory | Purpose |
|---|---|
| `depcheck/` | Verify that Go module dependencies are correct and no unexpected direct deps were added |
| `shellcheck/` | Run shellcheck linter over all shell scripts in the repo |
| `linkchecker/` | Verify that all hyperlinks in documentation are reachable |
