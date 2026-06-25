# cmd/kueuectl-docs/

Documentation generator for kueuectl commands. Produces Markdown documentation from cobra command definitions.

## Purpose

Generates the reference documentation for `kueuectl` commands published to the Kueue website. Each command, flag, and example is extracted from the cobra command tree and formatted as Markdown.

## Usage

```bash
kueuectl-docs generate --output-dir=site/content/en/docs/reference/kubectl-kueue/
```

## Sub-packages

| Package | Purpose |
|---|---|
| `generators/` | Custom Markdown generator (extends cobra's built-in) |
| `templates/` | Go templates for command page layout |
