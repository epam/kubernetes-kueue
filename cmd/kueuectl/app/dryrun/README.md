# cmd/kueuectl/app/dryrun/

Dry-run flag helper for kueuectl commands.

## Purpose

Provides a shared `DryRunStrategy` type and flag registration helper used by subcommands that support `--dry-run`. Abstracts the difference between client-side dry-run (no API call) and server-side dry-run (`DryRun: All` on the API request).

## Key Type

`DryRunStrategy` — enum with values `None`, `Client`, `Server`. Commands call `AddDryRunFlag(cmd)` during setup, then inspect the strategy before mutating resources.
