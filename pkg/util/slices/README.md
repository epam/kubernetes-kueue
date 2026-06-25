# pkg/util/slices/

Slice utility functions supplementing the standard library `slices` package.

## Key Functions

- `Map[T, U any](s []T, f func(T) U) []U` — transform a slice
- `Filter[T any](s []T, keep func(T) bool) []T` — filter by predicate
- `Contains[T comparable](s []T, v T) bool` — membership check
- `FindIndex[T any](s []T, pred func(T) bool) int` — find first matching index
- `MappedContains[T any, K comparable](s []T, key func(T) K, v K) bool`

## Note

Kueue's linting rules require using the `slices` package functions where equivalent standard library functions exist (e.g., `slices.Contains` instead of a hand-rolled loop). This package provides extensions not in the standard library.
