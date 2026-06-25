# test/integration/singlecluster/scheduler/delayedadmission/

Integration tests for delayed admission scenarios.

## What's tested

- Workloads that cannot immediately be admitted due to external admission check gates
- AdmissionCheck controller integration — workloads held until the check approves
- ProvisioningRequest integration: workload waits for a ProvisioningRequest to be fulfilled before pods are created
- Interaction between WaitForPodsReady and external admission checks
