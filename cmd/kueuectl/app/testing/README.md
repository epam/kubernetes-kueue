# cmd/kueuectl/app/testing/

Test helpers for kueuectl unit tests.

## Purpose

Provides a `FakeClientGetter` that returns a pre-configured fake Kubernetes client (via `k8s.io/client-go/kubernetes/fake` and the generated Kueue fake client) without needing a real cluster. Command unit tests use this to inject controlled API responses and assert on objects created or patched.

## Usage

```go
tgetter := &testing.FakeClientGetter{
    K8sClient: k8sfake.NewSimpleClientset(existingPod),
    KueueClient: kueuefake.NewSimpleClientset(existingWorkload),
}
cmd := list.NewListWorkloadCommand(tgetter)
```
