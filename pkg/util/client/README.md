# pkg/util/client/

Kubernetes client wrappers and helpers.

## Key Functions

- `WithFieldOwner(c, owner string) client.Client` — wraps a client to always use server-side apply with a specific field manager
- `IsOwner(wl, job)` — check ownership relationship
- `ListAllWithLabels(ctx, c, labels, list)` — list all objects matching labels across all namespaces

## Usage

Primarily used when Kueue needs to apply partial updates (server-side apply) without overwriting fields owned by other controllers.
