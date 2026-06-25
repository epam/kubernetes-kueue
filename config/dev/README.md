# config/dev/

Development overlay for deploying Kueue with debug settings enabled.

## Differences from `config/default/`

- Log verbosity raised to `V(6)` for detailed debug output
- Feature gates may be enabled that are not on by default
- Optimised for fast iteration on a local kind cluster, not for production use

## Usage

```bash
kubectl apply -k config/dev
```
