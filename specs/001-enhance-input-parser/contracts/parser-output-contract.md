# Contract: Parser Output and Execution Mapping

## Purpose

定義 parser 輸出到下游（task/scanapp/writer）的資料契約與去重語義。

## Output Records

### Accepted Parsed Target Record

Each accepted row MUST expose:

- destination tuple: `dst_ip`, `port`, `protocol`
- source context: `src_ip`, `src_network_segment`
- target context: `dst_network_segment`, `service_label`
- policy context: `decision`, `policy_id`, `reason`
- traceability key: `execution_key`

### Rejected Row Result

Each rejected row MUST expose:

- `row_index`
- rejection reason (machine-friendly code + human-readable message)

## Deduplication Contract

1. Execution dedup key MUST be `{dst_ip, port, protocol}` rendered as `dst_ip:port/protocol`.
2. Scanner execution MUST happen once per dedup key.
3. System MUST preserve mapping from each source row to its dedup execution key.
4. Distinct policy/background rows sharing same dedup key MUST remain traceable in output context fields (merged with `|` delimiter when needed).

## Writer Output Contract

Writer rows MUST keep existing leading columns and append rich context columns:

`ip, ip_cidr, port, status, response_time_ms, fab_name, cidr_name, service_label, decision, policy_id, reason, execution_key, src_ip, src_network_segment`

- Existing consumers depending on leading columns remain compatible.
- Rich fields MUST NOT be silently dropped.

## Summary Contract

After parsing, system MUST emit summary containing:

- total row count
- accepted row count
- rejected row count
- rejection reason buckets
- deduplicated execution target count
