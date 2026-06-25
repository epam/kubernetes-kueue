# cmd/kueuectl/app/flags/

Shared flag definitions for kueuectl subcommands.

## Purpose

Centralises flag name constants and registration helpers that are reused across multiple kueuectl commands (e.g., `--namespace`, `--selector`, `--field-selector`, `--output`, `--all-namespaces`). Keeps flag names consistent so tab-completion and documentation are uniform.

## Contents

- `flags.go` — flag registration helpers and constant strings for shared flag names
