# apis/config/v1beta1/

Deprecated v1beta1 configuration API for the Kueue controller manager. Kept for backwards compatibility with existing installations that reference this version.

## Status

**Deprecated.** New installations should use `v1beta2`. This version is retained only for migration support.

## Key Type: `Configuration`

Structurally similar to `v1beta2` but with fewer fields. Missing fields that were added in v1beta2 (e.g., `MultiKueue`, `Resources`, `FairSharing` details).

## Conversion

Automatic conversion to/from `v1beta2` is handled by `zz_generated.conversion.go`. Controllers always operate on `v1beta2` internally.
