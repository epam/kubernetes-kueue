# cmd/importer/pod/

Pod-specific import logic for the Kueue importer.

## Responsibilities

- Enumerate Pods matching the import selector
- Extract resource requests from each Pod
- Group Pods by their logical job (using owner references)
- Create `Workload` objects with pre-admitted status
- Add Kueue labels to Pods so they are tracked going forward

## Workload Creation

For each imported job:
```go
Workload{
    Spec: {
        QueueName: targetLocalQueue,
        PodSets: [{
            Count: podCount,
            Template: podTemplate,
        }],
    },
    Status: {
        Admission: {
            ClusterQueue: targetCQ,
            // ... pre-set as admitted
        },
    },
}
```

The pre-admitted status means Kueue accounts for these resources immediately without re-running them through the scheduler.
