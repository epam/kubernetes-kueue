# test/integration/singlecluster/scheduler/fairsharing/

Integration tests for Kueue's fair sharing and Dominant Resource Sharing (DRS) preemption.

## Purpose

Verifies that when multiple LocalQueues compete for shared ClusterQueue resources, the scheduler uses DRS share values to preempt workloads and reorder the queue fairly.

## What's tested

- DRS-based preemption: lower-share workload is preempted when a higher-priority tenant needs resources
- Admission Fair Sharing (AFS): workloads are admitted in DRS order, not FIFO
- DRS metrics: `kueue_cluster_queue_dominant_resource_share` correctness
- DRA fair sharing: DRA ResourceClaims participate in DRS computation
- FairSharing feature gate: behaviour is correct both when on and off
