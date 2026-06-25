# test/integration/singlecluster/controller/admissionchecks/

Integration tests for AdmissionCheck controllers.

## Sub-packages

| Package | What it tests |
|---|---|
| `provisioning/` | ProvisioningRequest AdmissionCheck controller — creates and tracks ProvisioningRequest objects, releases the AdmissionCheck gate when provisioning succeeds |

## Adding new AdmissionCheck tests

When adding a new AdmissionCheck controller (e.g., for a new external provisioner), create a new sub-package here following the pattern in `provisioning/`.
