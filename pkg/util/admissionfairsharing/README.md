# pkg/util/admissionfairsharing/

Admission Fair Sharing (AFS) calculation utilities. Computes submitter-level DRS values used for queue ordering when the `AdmissionFairSharing` feature gate is enabled.

## Key Functions

- `ComputeDRS(usage, total FlavorResourceQuantities) float64` — compute Dominant Resource Share
- `SubmitterUsage(wls []workload.Info) FlavorResourceQuantities` — aggregate usage for a submitter (group of workloads with the same submitter label)
- `OrderByShare(wls []workload.Info, totalCapacity)` — sort workloads by submitter DRS (ascending)

## DRS Computation

```
DRS(submitter) = max over resources { submitter_usage[r] / total_capacity[r] }
```

Submitters with lower DRS are placed earlier in the queue (they get priority admission to balance fairness).
