# pkg/config/

Configuration loading, validation, and defaulting for the Kueue controller manager.

## Responsibilities

- Load `Configuration` from a file path (YAML/JSON) at startup
- Apply defaults for unset fields
- Validate the configuration (unsupported field combinations, value ranges)
- Expose the final `Configuration` to other packages

## Key Functions

- `Load(path string) (*v1beta2.Configuration, error)` — decode and validate config from file
- `SetDefaults(cfg *v1beta2.Configuration)` — apply default values

## Configuration Sources

In order of precedence:
1. Command-line flag `--config=/path/to/config.yaml`
2. Default `Configuration` values

## Validation Rules

- `Integrations.Frameworks` must be valid API group strings
- `MultiKueue.GCInterval` must be positive
- `QueueVisibility.ClusterQueues.MaxCount` must be ≥ 1
- TLS profiles must reference valid cipher suites
