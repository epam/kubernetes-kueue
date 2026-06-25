# cmd/kueueviz/frontend/

React/TypeScript frontend for KueueViz — a web UI for visualizing Kueue workload and queue state.

## Technology Stack

- **React 18** — UI framework
- **TypeScript** — type safety
- **Vite** — build tooling
- **WebSocket** — real-time updates from backend

## Features

- ClusterQueue resource usage visualization (quota bars, utilization gauges)
- Workload list with status, priority, queue position
- LocalQueue drill-down view
- Real-time updates via WebSocket (no manual refresh needed)
- Responsive layout for various screen sizes

## Development

```bash
cd cmd/kueueviz/frontend
npm install
npm run dev      # start dev server with hot reload
npm run build    # production build → dist/
npm run test     # run unit tests
```

## Production Build

The frontend is built into static files embedded in the backend binary via Go's `embed` package. `make build-kueueviz` runs both `npm run build` and `go build ./cmd/kueueviz/...`.

## `src/` Structure

```
src/
├── App.tsx              # Root component
├── components/          # Reusable UI components
├── hooks/               # Custom React hooks (WebSocket, data fetching)
├── pages/               # Page-level components
├── types/               # TypeScript type definitions
└── utils/               # Utility functions
```
