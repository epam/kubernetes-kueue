# pkg/util/orderedgroups/

Ordered group data structure. A map-like structure that preserves insertion order for key-group pairs.

## Key Type: `OrderedGroups[K, V]`

```go
type OrderedGroups[K comparable, V any] struct { ... }

func (g *OrderedGroups[K, V]) Add(key K, value V)
func (g *OrderedGroups[K, V]) Get(key K) ([]V, bool)
func (g *OrderedGroups[K, V]) Keys() []K  // in insertion order
func (g *OrderedGroups[K, V]) ForEach(f func(K, []V))
```

## Usage

Used in the scheduler to maintain an ordered mapping from ClusterQueue to workload heads, preserving the iteration order for fair scheduling.
