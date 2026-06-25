# pkg/util/expectations/

Controller expectations tracking. Prevents controllers from acting on stale cache state when they have pending writes that haven't been reflected in the informer cache yet.

## Problem

After a controller creates/deletes an object, the informer may not have seen the change yet. A subsequent reconcile sees stale state and might duplicate the action.

## Solution

The `Expectations` tracker records pending operations:
```go
exp.ExpectCreations(key, n)  // expect n new objects to appear
exp.ExpectDeletions(key, n)  // expect n objects to disappear
exp.SatisfiedExpectations(key) bool  // safe to reconcile?
```

Controllers skip reconciliation if `!SatisfiedExpectations()`, waiting for the informer to catch up.

## TTL

Expectations expire after a configurable TTL (default: 5 minutes) to prevent deadlocks if events are missed.
