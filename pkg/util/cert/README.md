# pkg/util/cert/

TLS certificate management for Kueue's internal webhook server.

## Purpose

When `Configuration.InternalCertManagement.Enable = true`, Kueue generates and rotates its own TLS certificate for the webhook server without requiring cert-manager.

## Key Functions

- `GenerateSelfSignedCert(hosts []string) (cert, key []byte, err error)` — create a self-signed cert
- `RotateIfExpiring(certPath, keyPath string, daysLeft int) error` — rotate cert if expiry is near
- `InjectCABundle(ctx, c, webhookName, ca []byte) error` — patch the `ValidatingWebhookConfiguration`/`MutatingWebhookConfiguration` with the CA bundle

## Certificate Lifecycle

1. At startup, check if cert files exist and are valid
2. If not, generate new self-signed cert
3. Inject CA bundle into webhook configurations
4. Set up a background rotation check (default: rotate when < 30 days remaining)
