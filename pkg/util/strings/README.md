# pkg/util/strings/

String utility functions.

## Key Functions

- `MaybeShorten(s string, maxLen int) string` — truncate with ellipsis if too long
- `SanitizeLabelValue(s string) string` — make a string safe for use as a Kubernetes label value
- `JoinWithOr(items []string) string` — join with commas and "or" (for error messages)
