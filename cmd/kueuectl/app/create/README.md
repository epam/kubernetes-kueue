# cmd/kueuectl/app/create/

`kueuectl create` subcommands.

## Commands

### `kueuectl create localqueue <name>`

```bash
kueuectl create localqueue my-queue \
  --clusterqueue=team-cq \
  --namespace=my-namespace
```

Creates a `LocalQueue` with the specified name pointing to the given `ClusterQueue`.

### `kueuectl create clusterqueue <name>`

```bash
kueuectl create clusterqueue my-cq \
  --cohort=team-cohort \
  --nominal-quota=cpu=10,memory=100Gi \
  --resource-flavor=default-flavor
```

Creates a `ClusterQueue` with basic quota configuration.

## Options

Both commands support:
- `--dry-run=client|server` — preview without creating
- `-o yaml|json` — output format
