# cmd/kueuectl/app/version/

`kueuectl version` subcommand implementation.

## Purpose

Prints the kueuectl client version and, if reachable, the server-side Kueue controller version. The client version is embedded at build time via `pkg/version`. The server version is queried from the running Kueue controller manager's `/version` endpoint or from the `kueue-controller-manager` deployment's image tag.

## Output

```
Client Version: v0.9.0
Server Version: v0.9.0
```
