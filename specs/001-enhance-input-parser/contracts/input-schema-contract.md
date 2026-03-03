# Contract: Rich Input Schema for CIDR Traffic Rows

## Purpose

定義輸入檔的欄位契約與匹配規則，作為 parser 對外行為邊界。

## Canonical Headers (Required)

- `src_ip`
- `src_network_segment`
- `dst_ip`
- `dst_network_segment`
- `service_label`
- `protocol`
- `port`
- `decision`
- `policy_id`
- `reason`

## Header Matching Rules

1. Matching MUST use canonical names only.
2. Header comparison MUST ignore case and surrounding spaces.
3. Alias names (e.g., `source_ip`) MUST NOT be accepted.

## Row Validation Rules

1. All 10 canonical fields are mandatory per row.
2. `protocol` allowed values: `tcp` only.
3. `decision` allowed values: `accept`, `deny`.
4. `port` must be numeric and within 1..65535.
5. `src_ip` must be inside `src_network_segment`.
6. `dst_ip` must be inside `dst_network_segment`.

## Failure Contract

- Invalid rows MUST be rejected at row level.
- Parser MUST produce row-level failure reasons and aggregate summary.
- If all rows are invalid, parser MUST report no usable input and stop downstream flow.

## Validation Code Set

- `missing_field`
- `invalid_protocol`
- `invalid_decision`
- `invalid_port`
- `invalid_src_ip`
- `invalid_dst_ip`
- `invalid_src_network_segment`
- `invalid_dst_network_segment`
- `src_containment_mismatch`
- `dst_containment_mismatch`
