# Pressure API Contract

## Endpoint

- Method: `GET`
- URL: from `-pressure-api`
- Poll interval: from `-pressure-interval`

## Success Response

- HTTP status `< 400`
- JSON body includes `pressure` field
- Accepted `pressure` types:
  - number
  - integer
  - numeric string (parseable as integer)

Example:

```json
{"pressure": 72}
```

## Failure Conditions

Any of the following counts as one failed poll:
- HTTP status `>= 400`
- invalid JSON body
- missing `pressure` field
- unsupported `pressure` type
- network timeout/connection error

## Failure Escalation Policy

- Consecutive failures `1` and `2`: log error, keep last known pause state.
- Consecutive failure `3`: terminate scan as fatal and trigger resume save flow.

## Pause Semantics

- `api_paused = (pressure >= threshold)`
- Effective dispatch pause condition is OR-gate:
  - `api_paused || manual_paused`
- Resume dispatch only when both become false.
