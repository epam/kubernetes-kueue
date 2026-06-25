# pkg/util/maps/

Map utility functions. Supplements the standard library `maps` package with Kueue-specific helpers.

## Key Functions

- `Clone[K, V](m map[K]V) map[K]V` — deep clone a map
- `Merge[K, V](a, b map[K]V) map[K]V` — merge two maps (b overrides a)
- `Keys[K, V](m map[K]V) []K` — return sorted keys
- `FilterKeys[K, V](m map[K]V, keep func(K) bool) map[K]V` — filter by key predicate
