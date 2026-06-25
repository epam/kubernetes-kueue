# cmd/kueueviz/backend/middleware/

HTTP middleware for the KueueViz backend server.

## Middleware Stack

### Authentication Middleware

When KueueViz is deployed in-cluster with authentication enabled:
- Validates Bearer tokens via Kubernetes TokenReview API
- Extracts user identity for audit logging
- Returns 401 for invalid/expired tokens

### CORS Middleware

Handles Cross-Origin Resource Sharing for the React frontend:
- Allows requests from the configured frontend origin
- Sets appropriate headers for WebSocket upgrade requests

### Logging Middleware

Structured request logging:
- Request method, path, duration
- Response status code
- Client IP (from `X-Forwarded-For` or direct)

## Configuration

Middleware behavior is configured via `cmd/kueueviz/backend/config/`.
