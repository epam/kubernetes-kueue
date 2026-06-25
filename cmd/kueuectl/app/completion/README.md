# cmd/kueuectl/app/completion/

Shell completion support for `kueuectl`.

## Purpose

Registers the `completion` subcommand that generates shell completion scripts for bash, zsh, fish, and PowerShell. Users source the output into their shell session to enable tab-completion for kueuectl commands, flags, and resource names.

## Usage

```bash
source <(kueuectl completion bash)
kueuectl completion zsh > ~/.zsh/completions/_kueuectl
```

## Implementation

Thin wrapper around cobra's built-in completion generation (`cmd.GenBashCompletion`, etc.). No custom completion logic is needed because kueuectl uses the standard cobra completion framework.
