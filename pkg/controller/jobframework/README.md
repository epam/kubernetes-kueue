# pkg/controller/jobframework/

Plugin framework that defines the `GenericJob` interface and base reconciler used by all job type integrations. Every job adapter (batch/Job, RayJob, PyTorchJob, etc.) builds on this framework.

## Core Interface: `GenericJob`

```go
type GenericJob interface {
    Object() client.Object         // the job object itself
    IsSuspended() bool             // is the job currently suspended?
    Suspend()                      // suspend the job (called by Kueue during eviction)
    RunWithPodSetsInfo(ctx, c, podSetsInfo) error  // unsuspend + inject node affinity
    RestorePodSetsInfo(podSetsInfo) bool           // undo RunWithPodSetsInfo
    Finished(ctx) (msg string, success, finished bool)
    PodSets(ctx, c) ([]kueue.PodSet, error)        // describe resource requirements
    IsActive() bool                // are pods currently running?
    PodsReady(ctx, c) bool         // are all pods in Ready state?
    GVK() schema.GroupVersionKind
}
```

## Optional Interfaces

| Interface | Purpose |
|---|---|
| `JobWithPodLabelSelector` | Custom label selector for job pods |
| `JobWithReclaimablePods` | Support partial release of running pods |
| `JobWithCustomStop` | Custom stop/suspend behavior |
| `JobWithFinalizerSetup` | Custom finalizer management |
| `MultiKueueAdapter` | MultiKueue dispatch support |

## Base Reconciler (`reconciler.go`)

The `JobReconciler` handles the common lifecycle for all job types:

1. **Ensure workload exists** — create `Workload` object if missing
2. **Sync suspension** — if Kueue says suspend, call `job.Suspend()`
3. **Wait for admission** — if workload is admitted, call `job.RunWithPodSetsInfo()`
4. **Handle finish** — when job finishes, mark workload as `Finished`
5. **Handle eviction** — remove workload admission on eviction events

## Integration Registration (`integrationmanager.go`)

```go
// In each job adapter's init():
jobframework.RegisterIntegration("batch/v1", jobframework.IntegrationCallbacks{
    NewJob:            func() GenericJob { return &batchjob.Job{} },
    SetupIndexes:      batchjob.SetupIndexes,
    AddToScheme:       batchv1.AddToScheme,
    NewReconciler:     batchjob.NewReconciler,
    SetupWebhook:      batchjob.SetupWebhook,
    MultiKueueAdapter: &batchjob.MultiKueueAdapter{},
})
```

## Validation (`validation.go`)

Shared webhook validation logic for all job types:
- `queue-name` label is set to a valid LocalQueue
- Mutation constraints (cannot change queue while running)
- TAS validation (`tas_validation.go`)

## Event Reasons (`events.go`, `stop_reason.go`)

Defines constants for Kubernetes events emitted during job lifecycle:
- `StartedAdmission`, `FailedAdmission`, `Evicted`, `Preempted`
- `StopReason` enum: `Evicted`, `PreemptedByFairSharing`, `Cancelled`, etc.
