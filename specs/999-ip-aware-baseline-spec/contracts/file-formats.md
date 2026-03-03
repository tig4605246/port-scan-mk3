# File Format Contract

## 1) CIDR Input CSV

- Must include header row.
- Required logical columns are resolved by flag names:
  - `-cidr-ip-col` -> target selector column
  - `-cidr-ip-cidr-col` -> boundary CIDR column
- Column matching is **case-sensitive exact match**.
- Duplicate same-name header is fatal.
- Selector supports IPv4 single address or IPv4 CIDR.

Example:

```csv
team,ip,ip_cidr,note
fab1,192.168.1.10,192.168.1.0/24,edge-host
fab1,192.168.1.16/30,192.168.1.0/24,subrange
```

## 2) Port Input CSV

- No header.
- Each line follows `<port>/tcp`.
- Non-TCP lines are invalid.

Example:

```csv
22/tcp
80/tcp
443/tcp
```

## 3) scan_results CSV (per batch)

Header:

```csv
ip,ip_cidr,port,status,response_time_ms,fab_name,cidr_name
```

Status values:
- `open`
- `close`
- `close(timeout)`

## 4) opened_results CSV (per batch)

- Same header as `scan_results`.
- Must contain only rows where `status == open`.
- If no open rows, keep header only.

## 5) Resume State JSON

- JSON array of chunk progress entries.
- Required fields:
  - `cidr`
  - `ports`
  - `next_index`
  - `scanned_count`
  - `total_count`
  - `status`
- Used for restart without duplicate/missing tasks.
