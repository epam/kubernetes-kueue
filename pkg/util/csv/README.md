# pkg/util/csv/

CSV parsing utilities.

## Key Functions

- `ParseCSVMap(s string) map[string]string` — parse `"key1=val1,key2=val2"` into a map
- `FormatCSVMap(m map[string]string) string` — format a map as `"key1=val1,key2=val2"`

## Usage

Used for parsing feature gate flags and label selector strings from command-line arguments and configuration files.
