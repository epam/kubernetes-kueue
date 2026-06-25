# test/integration/singlecluster/controller/admissionchecks/provisioning/

Integration tests for the ProvisioningRequest AdmissionCheck controller.

## Purpose

Verifies the complete lifecycle of workloads gated by a `ProvisioningRequestConfig` AdmissionCheck, which requires Cluster Autoscaler to provision additional nodes before the workload is admitted.

## What's tested

- `ProvisioningRequest` object is created when a workload has a `ProvisioningRequestConfig` admission check
- Workload is held (AdmissionCheck pending) until the ProvisioningRequest is marked `Provisioned`
- After provisioning, the AdmissionCheck is released and the workload is admitted
- ProvisioningRequest is deleted when the workload finishes or is evicted
- Retry logic: a failed ProvisioningRequest triggers a new request after backoff
- `ProvisioningRequestConfig` parameters (max retries, backoff) are respected
