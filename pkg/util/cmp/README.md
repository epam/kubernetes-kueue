# pkg/util/cmp/

Comparison utilities for testing and production use.

## Key Functions

- `Diff(a, b interface{}, opts ...cmp.Option) string` — produce a human-readable diff between two values (wraps `go-cmp`)
- `DiffNoOrder[T comparable](a, b []T) string` — diff two slices ignoring element order
- `NoSortRequests(opts)` — `cmp.Option` for ignoring resource quantity sort order

## Usage

Primarily used in tests to generate readable diffs:
```go
if diff := utilcmp.Diff(expected, actual); diff != "" {
    t.Errorf("unexpected result:\n%s", diff)
}
```

Also used in controllers to check if a status update would produce an actual change.
