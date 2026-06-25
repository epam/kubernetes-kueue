# cmd/kueueviz/backend/handlers/

HTTP and WebSocket handlers for the KueueViz backend API.

## Handlers

### WebSocket Handler

Streams real-time updates of Kueue state to connected browser clients:
- ClusterQueue resource usage updates
- Workload status changes
- Queue depth changes

Uses Kubernetes informers under the hood — changes propagate from the API server to clients with minimal latency.

### REST Handlers

| Endpoint | Description |
|---|---|
| `GET /api/v1/clusterqueues` | List all ClusterQueues with usage |
| `GET /api/v1/localqueues` | List all LocalQueues |
| `GET /api/v1/workloads` | List workloads with filtering |
| `GET /api/v1/clusterqueues/{name}` | Single ClusterQueue detail |
| `GET /healthz` | Health check |

## Response Format

All REST endpoints return JSON. The WebSocket sends JSON events matching the REST response schema, tagged with an event type (`update`, `delete`, `add`).
