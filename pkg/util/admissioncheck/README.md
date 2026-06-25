# pkg/util/admissioncheck/

Utilities for working with `AdmissionCheck` status on `Workload` objects.

## Key Functions

- `FilterForController(wl, controllerName)` — return only the AdmissionCheck states owned by a specific controller
- `SetAdmissionCheckState(wl, name, state, msg)` — update a specific check's state on the workload
- `AllChecksPassed(wl)` — returns true when all required checks are `Ready`
- `FindAdmissionCheck(checks, name)` — find a specific check by name

## Usage

```go
// In a MultiKueue admission check controller:
acsUtil.SetAdmissionCheckState(&wl, "multikueue", kueue.CheckStateReady, "")
```
