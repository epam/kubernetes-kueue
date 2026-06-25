# pkg/util/podset/

Pod set utility functions for manipulating `PodSet` slices.

## Key Functions

- `FindPodSet(podSets []kueue.PodSet, name string) (*kueue.PodSet, bool)` — find a PodSet by name
- `TotalCount(podSets []kueue.PodSet) int32` — sum all pod counts
- `ValidatePodSets(podSets []kueue.PodSet) error` — check for duplicate names, invalid counts
- `AssignmentForPodSet(assignments []kueue.PodSetAssignment, name string)` — find an assignment by pod set name
