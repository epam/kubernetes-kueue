# cmd/kueueviz/frontend/src/

KueueViz React frontend source code.

## Key Components

- **ClusterQueueList** — table of all ClusterQueues with usage/quota bars
- **LocalQueueView** — detailed view for a single LocalQueue
- **WorkloadTable** — filterable/sortable workload list
- **ResourceUsageBar** — visual progress bar for resource utilization
- **WebSocketProvider** — React context providing real-time updates

## `utils/`

Utility functions including:
- WebSocket message parsing and event dispatch
- Resource quantity formatting (`"1.5 CPU"`, `"8Gi"`)
- Status color mapping (green=active, yellow=pending, red=error)
