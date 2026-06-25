# pkg/constants/

Project-wide constant values shared across multiple packages.

## Contents

- `KueueName` — the canonical name used in RBAC, labels, and annotations
- `ManagedBy` label value for Kueue-created objects
- Default timeouts, intervals, and size limits
- Common annotation/label key prefixes

## Usage

Import this package to avoid hardcoding string constants:
```go
import "sigs.k8s.io/kueue/pkg/constants"

label := constants.QueueLabel  // "kueue.x-k8s.io/queue-name"
```

## Note

Job-type-specific constants (pod integration constants, label keys) live in their respective packages (e.g., `pkg/controller/jobs/pod/constants/`). This package only contains constants shared across 3+ packages.
