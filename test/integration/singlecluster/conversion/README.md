# test/integration/singlecluster/conversion/

Integration tests for CRD conversion webhooks.

## Purpose

Verifies that objects created with the `v1beta1` API are correctly converted to `v1beta2` (and back) by the conversion webhook. This is critical for rolling upgrades where old and new controller versions coexist.

## What's tested

- Round-trip conversion: `v1beta1 → v1beta2 → v1beta1` preserves all fields
- Hub version (`v1beta1`) storage semantics
- Unknown fields are preserved through conversion (no data loss)
- ClusterQueue, LocalQueue, Workload, ResourceFlavor conversion
