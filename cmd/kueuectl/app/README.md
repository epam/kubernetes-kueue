# cmd/kueuectl/app/

Root command structure and all kueuectl subcommand implementations.

## Structure

```
app/
├── root.go              # Root cobra command, global flags
├── clientgetter/        # Kubernetes client factory
├── completion/          # Shell completion setup
├── create/              # kueuectl create subcommands
├── delete/              # kueuectl delete subcommands
├── dryrun/              # Dry-run support
├── flags/               # Shared flag definitions
├── list/                # kueuectl list subcommands
├── options/             # Global options (namespace, kubeconfig)
├── passthrough/         # Pass-through to kubectl
├── resume/              # kueuectl resume subcommands
├── stop/                # kueuectl stop subcommands
├── testing/             # Test helpers for kueuectl tests
└── version/             # kueuectl version command
```

## Root Command

```go
func NewKueuectlCmd(streams genericiooptions.IOStreams) *cobra.Command {
    cmd := &cobra.Command{
        Use: "kubectl-kueue",
    }
    cmd.AddCommand(
        create.NewCreateCmd(...),
        list.NewListCmd(...),
        stop.NewStopCmd(...),
        resume.NewResumeCmd(...),
        // ...
    )
    return cmd
}
```

## Global Flags

- `--namespace` / `-n` — target namespace
- `--kubeconfig` — path to kubeconfig
- `--context` — kubeconfig context
