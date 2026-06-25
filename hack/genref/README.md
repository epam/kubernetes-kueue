# hack/genref/

Reference documentation generator for the Kueue API types.

## Purpose

Generates the API reference documentation published to the Kueue website. Uses `gen-crd-api-reference-docs` (or equivalent) to extract Go type comments from the `apis/` tree and render them as Markdown.

## Files

| File | Purpose |
|---|---|
| `config.yaml` | Generator configuration — lists the API packages to document and output paths |
| `markdown/members.tpl` | Go template for rendering struct field documentation |
| `markdown/pkg.tpl` | Go template for rendering a package-level API reference page |
| `markdown/type.tpl` | Go template for rendering a single type (struct, interface, or type alias) |

## Usage

```bash
make generate-apiref
```

Output is written to the website documentation directory.
