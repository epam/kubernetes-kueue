# pkg/cache/queue/

Manages the ordered set of pending workloads waiting to be admitted. This is the "queue" in "job queuing system" — workloads enter here when submitted and leave when admitted or evicted.

## Key Type: `Manager`

The `Manager` owns all `LocalQueue` and `ClusterQueue` queue state. Controllers call it to:
- `AddOrUpdateWorkload(wl)` — enqueue a workload
- `RequeueWorkload(wl, reason)` — put an evicted workload back in the queue
- `DeleteWorkload(wl)` — remove from queue (on admission or deletion)
- `Heads()` — return the next batch of workloads ready for the scheduler to try

### `Heads()`

Returns one "head" workload per `ClusterQueue`. A head is the workload with the highest effective priority in that CQ's queue. The scheduler tries to admit all heads in one cycle.

## Ordering Strategies

The queue ordering is configurable per `ClusterQueue`:

| Strategy | Description |
|---|---|
| `BestEffortFIFO` | Admit as many as possible; FIFO within same priority |
| `StrictFIFO` | Never skip ahead — a high-resource workload blocks lower ones |

Fair sharing uses the `afs/` subpackage for ordering when `AdmissionFairSharing` is enabled.

## Sub-packages

| Package | Purpose |
|---|---|
| [`afs/`](afs/) | Admission FairSharing ordering: sort by DRS share to prevent starvation |
