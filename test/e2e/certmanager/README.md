# test/e2e/certmanager/

End-to-end tests for the cert-manager TLS integration.

## Purpose

Verifies that when `--certManagerEnabled=true` is set in the Kueue Helm chart, cert-manager correctly provisions TLS certificates for:
- Webhook servers (validating + mutating)
- Visibility API server

## Tests

- `certmanager_test.go` — cert-manager Certificate objects are created and ready
- `visibility_test.go` — Visibility API server is reachable with cert-manager-issued TLS
- `metrics_test.go` — metrics endpoint TLS with cert-manager certificate

## Prerequisites

cert-manager must be installed in the cluster before running these tests:
```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml
```
