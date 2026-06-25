# cmd/kueuectl/app/options/

Shared options structs for kueuectl subcommands.

## Purpose

Holds the `UpdateWorkloadActivationOptions` struct used by both `stop workload` and `resume workload`. Centralising the options type avoids duplication between two commands that operate on the same workload activation field.

## Key Type

`UpdateWorkloadActivationOptions` — contains the workload name, namespace, and the target `active` boolean. Both `stop` and `resume` embed this struct and only override the `active` value they set.
