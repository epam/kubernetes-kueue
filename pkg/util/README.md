# pkg/util/

Shared utility packages used throughout Kueue. Each sub-package is a small, focused library with no Kueue-specific business logic — they could theoretically be used in any Go project.

## Sub-packages

| Package | Purpose |
|---|---|
| [`admissioncheck/`](admissioncheck/) | AdmissionCheck status helpers |
| [`admissionfairsharing/`](admissionfairsharing/) | Admission fair sharing calculations |
| [`api/`](api/) | Kubernetes API helpers (patch, update) |
| [`cert/`](cert/) | TLS certificate management |
| [`client/`](client/) | Kubernetes client wrappers |
| [`cmp/`](cmp/) | Comparison utilities |
| [`csv/`](csv/) | CSV parsing |
| [`equality/`](equality/) | Deep equality helpers |
| [`expectations/`](expectations/) | Controller expectations tracking |
| [`heap/`](heap/) | Generic heap data structure |
| [`kubeversion/`](kubeversion/) | Kubernetes version detection |
| [`limitrange/`](limitrange/) | LimitRange constraint checking |
| [`logging/`](logging/) | Structured logging helpers |
| [`maps/`](maps/) | Map utilities |
| [`math/`](math/) | Math/rounding helpers |
| [`orderedgroups/`](orderedgroups/) | Ordered group data structure |
| [`parallelize/`](parallelize/) | Parallel execution helpers |
| [`pod/`](pod/) | Pod utilities |
| [`podset/`](podset/) | Pod set utilities |
| [`priority/`](priority/) | Workload priority helpers |
| [`ptr/`](ptr/) | Pointer utilities |
| [`queue/`](queue/) | Queue utilities |
| [`resource/`](resource/) | Resource quantity helpers |
| [`roletracker/`](roletracker/) | RBAC role tracking |
| [`routine/`](routine/) | Goroutine utilities |
| [`slices/`](slices/) | Slice utilities |
| [`statefulset/`](statefulset/) | StatefulSet helpers |
| [`strings/`](strings/) | String utilities |
| [`taints/`](taints/) | Node taint utilities |
| [`tas/`](tas/) | TAS utilities |
| [`testing/`](testing/) | Test builder/fixture utilities |
| [`testingjobs/`](testingjobs/) | Per-framework test job builders |
| [`tlsconfig/`](tlsconfig/) | TLS configuration helpers |
| [`tolerations/`](tolerations/) | Node toleration utilities |
| [`useragent/`](useragent/) | HTTP user agent string |
| [`wait/`](wait/) | Wait/retry utilities |
| [`waitforpodsready/`](waitforpodsready/) | Wait-for-pods-ready logic |
| [`webhook/`](webhook/) | Webhook utility helpers |
