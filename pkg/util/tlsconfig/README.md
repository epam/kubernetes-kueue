# pkg/util/tlsconfig/

TLS configuration helpers. Translates Kueue's `TLSConfig` API type into Go's `tls.Config`.

## Key Functions

- `ToTLSConfig(cfg *v1beta2.TLSConfig) (*tls.Config, error)` — build a `tls.Config` from the Kueue configuration
- `ApplyProfile(cfg *tls.Config, profile string) error` — apply a named security profile (Intermediate, Modern, Old) as defined by Mozilla TLS guidelines

## Profiles

| Profile | Min TLS | Cipher Suites |
|---|---|---|
| `Old` | TLS 1.0 | Broad compatibility |
| `Intermediate` | TLS 1.2 | Recommended default |
| `Modern` | TLS 1.3 | Strongest security |

Used by the webhook server and metrics server TLS configuration.
