# pkg/cache/hierarchy/

Implements the hierarchical cohort tree used for quota borrowing and lending between ClusterQueues.

## Concept

ClusterQueues can belong to a `Cohort`. Cohorts can themselves have parent cohorts, forming a tree. A workload in a ClusterQueue can borrow unused quota from any ClusterQueue in the same subtree, subject to `borrowingLimit` constraints.

```
Cohort: org
├── Cohort: team-a
│   ├── ClusterQueue: cq-a1
│   └── ClusterQueue: cq-a2
└── ClusterQueue: cq-b
```

## Key Types

### `ClusterQueueNode`

Represents a `ClusterQueue` in the hierarchy. Tracks:
- Parent cohort reference
- Own nominal quota (guaranteed resources)
- Lending limit (max it can lend to siblings)
- Borrowing limit per flavor

### `CohortNode`

Represents a `Cohort` in the hierarchy. Tracks:
- Parent cohort reference
- Sum of all descendant ClusterQueue nominal quotas (for aggregating lendable quota)
- Children (mix of CohortNodes and ClusterQueueNodes)

## Quota Propagation

When computing how much a ClusterQueue can borrow:
1. Walk up the tree aggregating unused quota from siblings at each level
2. Apply `lendingLimit` constraints at each node
3. Apply `borrowingLimit` constraint at the requesting ClusterQueue

This tree traversal happens inside the scheduler snapshot during flavor assignment.
