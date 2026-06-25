# cmd/kueue/

Main entry point for the Kueue controller manager binary.

## Responsibilities

`main.go` performs the startup sequence:

1. Parse flags (`--config`, `--feature-gates`, `--leader-elect`, etc.)
2. Load `Configuration` from file via `pkg/config`
3. Initialize feature gates via `pkg/features`
4. Create `controller-runtime` manager with scheme, metrics, health probes, leader election
5. Register all integrations (`jobframework.RegisterIntegration()` in each adapter's `init()`)
6. Setup core controllers (`pkg/controller/core`)
7. Setup job controllers for each enabled integration
8. Setup webhooks (`pkg/webhooks`)
9. Setup visibility server (`pkg/visibility`)
10. Start the scheduler goroutine (`pkg/scheduler`)
11. Start the manager (blocks until signal)

## Key Flags

| Flag | Default | Description |
|---|---|---|
| `--config` | — | Path to `Configuration` file |
| `--leader-elect` | `false` | Enable leader election |
| `--feature-gates` | — | Override feature gate defaults |
| `--metrics-bind-address` | `:8080` | Metrics endpoint |
| `--health-probe-bind-address` | `:8081` | Health/readiness probe endpoint |
| `--webhook-port` | `9443` | Webhook server port |

## Container Entrypoint

```dockerfile
ENTRYPOINT ["/manager"]
```

The binary is named `/manager` in the container image.
