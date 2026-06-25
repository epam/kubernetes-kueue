# pkg/version/

Version information for the Kueue binary. Provides the version string embedded at build time.

## Contents

- `Version` — semver string (e.g., `v0.9.0`), set via `-ldflags` during build
- `GitCommit` — git commit SHA
- `BuildDate` — build timestamp

## Usage

```go
import "sigs.k8s.io/kueue/pkg/version"

fmt.Printf("Kueue %s (%s)\n", version.Version, version.GitCommit)
```

Used by `kueuectl version` and the controller manager startup log.
