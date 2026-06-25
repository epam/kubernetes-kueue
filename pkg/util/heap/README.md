# pkg/util/heap/

Generic heap (priority queue) data structure using Go generics.

## Key Type: `Heap[T]`

```go
type Heap[T any] struct {
    // min-heap by default; caller provides less-than comparison
}

func New[T any](less func(a, b T) bool) *Heap[T]

func (h *Heap[T]) Push(item T)
func (h *Heap[T]) Pop() (T, bool)
func (h *Heap[T]) Peek() (T, bool)
func (h *Heap[T]) Len() int
```

## Usage

Used by the queue manager to maintain ordered sets of pending workloads:
```go
q := heap.New[*workload.Info](func(a, b *workload.Info) bool {
    return a.Obj.Spec.Priority > b.Obj.Spec.Priority  // max-heap by priority
})
```
