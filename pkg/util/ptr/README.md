# pkg/util/ptr/

Pointer utility functions.

## Key Functions

- `To[T any](v T) *T` — take the address of a value (useful for literals: `ptr.To(int32(3))`)
- `Deref[T any](p *T, def T) T` — dereference with default if nil
- `Equal[T comparable](a, b *T) bool` — pointer-safe equality

## Usage

```go
job.Spec.Parallelism = ptr.To(int32(8))
count := ptr.Deref(job.Spec.Parallelism, 1)
```

Common pattern when working with Kubernetes API types that use pointer fields for optional values.
