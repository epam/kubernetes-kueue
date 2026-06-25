# pkg/util/queue/

Queue utility helpers for the workload admission queue.

## Key Functions

- `Key(wl *kueue.Workload) string` — compute the cache key for a workload (`namespace/name`)
- `KeyFromUID(uid types.UID) string` — key from workload UID
- `IsActive(wl) bool` — is the workload in an active queue state

## Usage

Used by `pkg/cache/queue/Manager` for consistent key computation across the queue and cache maps.
