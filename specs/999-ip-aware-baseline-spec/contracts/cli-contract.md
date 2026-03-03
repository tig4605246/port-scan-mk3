# CLI Contract

## Command Surface

```text
port-scan validate -cidr-file <path> -port-file <path> [flags]
port-scan scan     -cidr-file <path> -port-file <path> [flags]
```

## Common Flags

- `-cidr-file` (required)
- `-port-file` (required)
- `-cidr-ip-col` (default: `ip`, case-sensitive exact match)
- `-cidr-ip-cidr-col` (default: `ip_cidr`, case-sensitive exact match)
- `-output` (default base name: `scan_results.csv`; batch output uses timestamp naming policy)
- `-resume` (optional state path; if provided, same path used for read + subsequent writes)
- `-disable-api`
- `-pressure-api`
- `-pressure-interval`
- `-bucket-rate`
- `-bucket-capacity`
- `-workers`
- `-timeout`
- `-delay`
- `-log-level`
- `-format` (`human|json`)

## Exit Codes

- `0`: success
- `1`: runtime/validation failure in command flow
- `2`: argument/usage/config parsing error
- `130`: scan canceled by context/SIGINT

## validate Output Contract

### human format

```text
valid=<true|false> detail=<message>
```

### json format

```json
{
  "valid": true,
  "detail": "ok"
}
```

## scan Artifact Contract

For each scan run, output batch file names are UTC timestamped:

- main: `scan_results-YYYYMMDDTHHMMSSZ.csv`
- open-only: `opened_results-YYYYMMDDTHHMMSSZ.csv`

If collision occurs within the same second, append shared incremental suffix:

- `scan_results-YYYYMMDDTHHMMSSZ-<n>.csv`
- `opened_results-YYYYMMDDTHHMMSSZ-<n>.csv`

`<n>` is a positive integer and must match across the two files of the same batch.

Resume state file:
- with `-resume <state-path>`: read/write same `<state-path>`
- without `-resume`: `<output-dir>/resume_state.json`
