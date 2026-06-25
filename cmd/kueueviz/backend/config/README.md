# cmd/kueueviz/backend/config/

Configuration types and loading for the KueueViz backend server.

## Configuration

```go
type Config struct {
    Port             int           // HTTP server port (default: 8082)
    KubeConfigPath   string        // Path to kubeconfig (empty = in-cluster)
    AllowedOrigins   []string      // CORS allowed origins for frontend
    AuthMode         AuthMode      // None / KubernetesToken
    TLSCertFile      string        // TLS cert for HTTPS
    TLSKeyFile       string        // TLS key for HTTPS
}
```

## Loading

Configuration is loaded from environment variables and command-line flags at startup.
