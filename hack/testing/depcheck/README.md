# hack/testing/depcheck/

Dependency check verification for the Kueue Go module.

## Purpose

Runs `verify.sh` to ensure that:
- No unexpected new direct dependencies were added to `go.mod`
- All direct dependencies have corresponding entries in `go.sum`
- Vendored dependencies match `go.mod` (if vendoring is used)

## Running

```bash
make verify-vendor
# or directly
hack/testing/depcheck/verify.sh
```

This check runs in CI on every PR to prevent accidental dependency additions.
