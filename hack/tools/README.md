# hack/tools/

Development tools and code generation scripts.

## Sub-directories

| Directory | Purpose |
|---|---|
| `code-generator/` | Runs `controller-gen` and Kubernetes code-generator tools to regenerate CRD manifests, deepcopy, clientsets, listers, and informers |
| `compatibility-lifecycle/` | Generates the feature-gate compatibility lifecycle table in the docs from the gate definitions in `pkg/features/` |
| `mdtoc/` | Generates markdown table-of-contents for documentation files |
| `yaml-processor/` | A Go tool that applies structured YAML transformations (field patches, insertions) during the release manifest assembly |
| `metricsdoc/` | Generates the metrics reference documentation from registered Prometheus metrics |
| `ginkgo-top/` | Parses Ginkgo JSON output and summarises the slowest tests (useful for CI analysis) |
| `prow-runtimes/` | Python script that queries Prow CI run times across test jobs |

## Files

| File | Purpose |
|---|---|
| `go.mod` / `go.sum` | Separate Go module for tooling dependencies (keeps tool deps out of the main module) |
| `pinversion.go` | Validates that tool versions match the pinned versions in the CI configuration |
