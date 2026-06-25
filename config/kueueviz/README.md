# config/kueueviz/

Kustomize overlay for the KueueViz dashboard.

## Purpose

Adds KueueViz-specific patches on top of a base installation: Ingress host configuration, frontend ConfigMap with the backend URL, and any frontend-specific environment overrides.

## Usage

Apply on top of the default installation:

```bash
kubectl apply -k config/default
kubectl apply -k config/kueueviz
```

Or compose in a custom overlay:

```yaml
# kustomization.yaml
bases:
  - ../default
components:
  - ../components/kueueviz
patches:
  - path: kueueviz/frontend_configmap_patch.yaml
  - path: kueueviz/frontend_ingress_patch.yaml
  - path: kueueviz/backend_ingress_patch.yaml
```
