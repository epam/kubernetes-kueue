# pkg/debugger/

Runtime debugging and introspection utilities. Provides tooling for inspecting Kueue's internal state without requiring log-level changes.

## Contents

- In-process snapshot dumping — dump the current cache state to a debug endpoint or log
- Queue state inspection — show pending workloads and their queue positions
- Goroutine profiling helpers

## When to Use

Primarily used during development and debugging. In production, use the Visibility API (`pkg/visibility/`) or metrics for observability. The debugger provides lower-level internal state that is not exposed via the public API.
