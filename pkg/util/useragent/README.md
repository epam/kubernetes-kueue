# pkg/util/useragent/

HTTP User-Agent string construction for Kueue clients.

## Key Functions

- `Default() string` — returns `"kueue/v{version} ({os}/{arch})"`
- `ForComponent(component string) string` — returns `"kueue/{component}/v{version}"`

## Usage

Set on HTTP clients making requests to the Kubernetes API server so that API server logs and audit records identify Kueue as the client:

```go
cfg.UserAgent = useragent.Default()
```
