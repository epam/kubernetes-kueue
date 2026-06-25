# pkg/

Core Kueue implementation packages. All production code (excluding binaries) lives here.

## Package Overview

| Package | Purpose |
|---|---|
| [`cache/`](cache/) | In-memory state store for the scheduler |
| [`config/`](config/) | Configuration loading and validation |
| [`constants/`](constants/) | Shared constant values |
| [`controller/`](controller/) | All controller-runtime reconcilers |
| [`debugger/`](debugger/) | Runtime debugging and introspection |
| [`dra/`](dra/) | Dynamic Resource Allocation integration |
| [`features/`](features/) | Feature gate registry |
| [`metrics/`](metrics/) | Prometheus metrics definitions |
| [`podset/`](podset/) | PodSet types and operations |
| [`resources/`](resources/) | Resource quantity math utilities |
| [`scheduler/`](scheduler/) | Admission scheduling loop |
| [`util/`](util/) | Shared utility packages |
| [`version/`](version/) | Version string |
| [`visibility/`](visibility/) | Visibility API server |
| [`webhooks/`](webhooks/) | Mutating/validating admission webhooks |
| [`workload/`](workload/) | Workload abstraction layer |
| [`workloadslicing/`](workloadslicing/) | Elastic job workload slicing |

## Architecture

```
cmd/kueue/main.go
  ├── pkg/config         (load Configuration)
  ├── pkg/features       (register feature gates)
  ├── pkg/cache          (start in-memory store)
  ├── pkg/controller/*   (start all reconcilers)
  ├── pkg/scheduler      (start scheduling loop)
  ├── pkg/webhooks       (register admission webhooks)
  └── pkg/visibility     (register aggregated API)
```
