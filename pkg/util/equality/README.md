# pkg/util/equality/

Deep equality helpers for Kueue API types. Extends standard `reflect.DeepEqual` with Kubernetes-specific semantics.

## Key Functions

- `SemanticDeepEqual(a, b interface{}) bool` — equality that ignores irrelevant differences (e.g., nil vs. empty slice in API types)
- `ConditionsEqual(a, b []metav1.Condition) bool` — compare status conditions ignoring `LastTransitionTime`

## Why Not `reflect.DeepEqual`?

Kubernetes API objects have fields that are semantically equivalent but structurally different:
- `nil` vs. `[]string{}` (both mean "no items")
- `metav1.Condition.LastTransitionTime` changes every reconcile — you don't want to trigger updates based on timestamp alone
