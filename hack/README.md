# hack/

Development and CI automation scripts. Not part of the released binary — these are build, test, and release tooling for contributors.

## Top-level Scripts

| Script | Purpose |
|---|---|
| `utils.sh` | Shared shell utilities sourced by other scripts (color output, retry, cluster helpers) |
| `multiplatform-build.sh` | Cross-compile all binaries for linux/amd64, linux/arm64, darwin/amd64 |
| `bump-test-infra-k8s-version.sh` | Update the Kubernetes version pin in test infrastructure |
| `cherry_pick_pull.sh` | Cherry-pick a PR to a release branch |
| `migrate-to-v1beta2.sh` | Migrate existing cluster resources from v1beta1 to v1beta2 API |
| `helm-chart-package.sh` | Package the Helm chart for release |
| `dump_cache.sh` | Dump the in-memory cache state from a running controller (debug) |

## Sub-directories

| Directory | Purpose |
|---|---|
| `testing/` | E2E and integration test runner scripts, performance test scripts, linting |
| `tools/` | Code generation scripts (client-gen, deepcopy-gen, compatibility lifecycle) |
| `releasing/` | Release automation (stage, promote, generate krew manifest, sync release notes) |
| `genref/` | Reference documentation generation for kueuectl |
| `debugpod/` | Debug pod manifests for live cluster debugging |
