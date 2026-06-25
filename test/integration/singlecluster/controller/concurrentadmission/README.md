# test/integration/singlecluster/controller/concurrentadmission/

Integration tests for the ConcurrentAdmission feature.

## Purpose

ConcurrentAdmission allows a workload to simultaneously pursue multiple resource flavor assignments in parallel (speculative execution), rather than trying one flavor at a time.

## What's tested

- Multiple flavors are speculatively assigned to a single workload
- The first flavor that gets fully provisioned wins and the others are cancelled
- Resource quota is correctly accounted during speculative assignments
- Parent/child workload relationships are created and cleaned up correctly
- The feature gate (`ConcurrentAdmission`) enables and disables the behavior
