# pkg/util/roletracker/

RBAC role tracking for the WorkloadDispatcher feature. Tracks which users/service accounts have access to which LocalQueues.

## Purpose

When auto-routing workloads to queues based on submitter identity, Kueue needs to know which queues a given service account can access. This package maintains a cache of RBAC role bindings to LocalQueue resources.

## Key Type: `RoleTracker`

```go
type RoleTracker struct {
    // Maps (namespace, service account) → accessible LocalQueues
}

func (rt *RoleTracker) QueuesFor(ns, sa string) []string
func (rt *RoleTracker) OnRoleBindingChange(rb *rbacv1.RoleBinding)
```

## Usage

Used exclusively by `pkg/controller/workloaddispatcher/` when the `WorkloadDispatcher` feature gate is enabled.
