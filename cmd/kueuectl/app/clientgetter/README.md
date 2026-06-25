# cmd/kueuectl/app/clientgetter/

Kubernetes client factory for kueuectl. Creates typed and dynamic clients from kubeconfig for use by all kueuectl commands.

## Key Type: `ClientGetter`

```go
type ClientGetter interface {
    K8sClient() (kubernetes.Interface, error)
    KueueClient() (versioned.Interface, error)
    DynamicClient() (dynamic.Interface, error)
    RESTConfig() (*rest.Config, error)
}
```

Wraps `k8s.io/cli-runtime/pkg/genericclioptions.ConfigFlags` to provide consistent client initialization across all subcommands.
