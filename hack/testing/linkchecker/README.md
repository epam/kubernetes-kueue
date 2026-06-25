# hack/testing/linkchecker/

Hyperlink verification for Kueue documentation.

## Purpose

`verify.sh` checks that all hyperlinks in Markdown documentation files resolve to valid URLs. Catches broken external links and incorrect relative paths.

## Running

```bash
make verify-links
# or directly
hack/testing/linkchecker/verify.sh
```

External links are checked with HTTP HEAD requests. The check is rate-limited to avoid triggering external site rate limits.
