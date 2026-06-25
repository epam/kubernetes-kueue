# config/alpha-enabled/

Kustomize overlay that enables all alpha feature gates.

## Purpose

Used for testing and development when you want to enable all alpha features at once. The overlay patches `controller_manager_config.yaml` to set `featureGates` for every gate that is in `Alpha` status.

## Usage

```bash
kubectl apply -k config/alpha-enabled
```

**Warning:** Do not use in production. Alpha features have no stability guarantees and may change or be removed in any release.
