# pkg/util/routine/

Goroutine management utilities.

## Key Functions

- `RunWithContext(ctx context.Context, fn func(context.Context)) error` — run a function in a goroutine, propagate context cancellation
- `RunWithRecover(fn func(), panicHandler func(interface{}))` — run with panic recovery
- `WorkerPool(ctx, size int, work <-chan func()) error` — manage a fixed-size worker pool

## Usage

The scheduler loop uses `routine.RunWithContext` to ensure the scheduling goroutine exits cleanly when the controller manager is shut down.
