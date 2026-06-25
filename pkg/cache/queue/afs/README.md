# pkg/cache/queue/afs/

Admission Fair Sharing (AFS) queue ordering algorithm. When `AdmissionFairSharing` is enabled, this package determines the order in which pending workloads are presented to the scheduler, ensuring equitable resource distribution among competing submitters.

## Algorithm

AFS orders workloads by their submitter's current DRS (Dominant Resource Share). A submitter with lower current usage gets higher priority in the queue, preventing any single user from monopolizing resources.

The DRS for a submitter is computed as:
```
DRS = max over resources of (usage / total_capacity)
```

Workloads from a submitter with lower DRS are offered to the scheduler first.

## Integration

AFS ordering is applied inside `pkg/cache/queue/Manager.Heads()` when the `AdmissionFairSharing` feature gate is enabled. It replaces the default FIFO ordering within the same priority band.

## Configuration

Enabled via `Configuration.AdmissionFairSharing` in `apis/config/v1beta2/`.
