# hack/testing/shellcheck/

Shell script linting using `shellcheck`.

## Purpose

`verify.sh` runs `shellcheck` over all `.sh` files in the repository to catch common shell scripting errors, quoting issues, and portability problems.

## Running

```bash
make verify-shellcheck
# or directly
hack/testing/shellcheck/verify.sh
```
