# test/e2e/upgrade/

End-to-end tests for rolling upgrade validation.

## Purpose

Verifies that upgrading the Kueue controller manager from a previous version to the current version does not break existing workloads:
- Admitted workloads continue running through the upgrade
- Pending workloads are re-admitted correctly after the new version starts
- Finalizers left by the old version are handled by the new version
- CRD conversions work correctly (v1beta1 ↔ v1beta2)

## Test

`upgrade_validation_test.go` — installs the previous release, creates workloads, upgrades to the current build, and verifies no workloads are lost or stuck.
