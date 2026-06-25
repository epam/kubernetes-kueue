# pkg/util/math/

Math and rounding utilities for resource quantity calculations.

## Key Functions

- `Ceil(a, b int64) int64` — integer ceiling division (`(a + b - 1) / b`)
- `Min[T constraints.Ordered](a, b T) T` — minimum of two values
- `Max[T constraints.Ordered](a, b T) T` — maximum of two values
- `Clamp[T constraints.Ordered](v, lo, hi T) T` — clamp a value to [lo, hi]

## Usage

Used in resource calculations where integer ceiling division is needed (e.g., rounding up pod counts for partial admission).
