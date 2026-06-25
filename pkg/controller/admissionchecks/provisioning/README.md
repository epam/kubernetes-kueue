# pkg/controller/admissionchecks/provisioning/

`ProvisioningRequest` admission check controller. Integrates Kueue with Kubernetes Cluster Autoscaler's [ProvisioningRequest API](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/expander/provisioningrequest/apis/autoscaling.x-k8s.io/v1beta1/types.go) to request node provisioning before admitting a workload.

## Use Case

When a workload needs GPU nodes that don't currently exist, Kueue can request Cluster Autoscaler to provision them before running the job. This prevents workloads from being admitted to quota but then pending forever waiting for nodes.

## Flow

```
1. Workload admitted to quota (QuotaReserved)
2. ProvisioningRequest admission check: Pending
3. Kueue creates a ProvisioningRequest object
4. Cluster Autoscaler provisions nodes
5. ProvisioningRequest.status = Provisioned
6. Admission check: Ready
7. Job unsuspended, pods scheduled on new nodes
```

## Configuration

`ProvisioningRequestConfig` CRD specifies:
```go
type ProvisioningRequestConfigSpec struct {
    ProvisioningClassName string             // e.g. "check-capacity.autoscaling.x-k8s.io"
    Parameters           map[string]Parameter // Passed to Cluster Autoscaler
    RetryStrategy        RetryStrategy        // Backoff on failure
}
```

## Retry

If provisioning fails, the controller retries with exponential backoff per `RetryStrategy`. After max retries, the workload is evicted and requeued.
