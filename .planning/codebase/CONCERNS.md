# CONCERNS.md - Technical Concerns

## Current Issues

### LSP Errors

The following test files have compilation errors:
- `pkg/scanapp/group_builder_test.go` - `g.port undefined`
- `pkg/scanapp/scan_helpers_test.go` - `got.port undefined`

These are pre-existing test issues in the codebase.

## Technical Debt

1. **Test Maintenance**: Some test files have errors that prevent compilation
2. **Rich Dashboard**: Planned feature (SPEC-13) not yet implemented

## Known Limitations

- **IPv4 Only**: No IPv6 support
- **TCP Only**: UDP not supported
- **Pressure Threshold**: Hardcoded to 90 (not exposed as CLI flag)

## Security

- **No scanning of external networks**: E2E tests use isolated Docker networks
- **No secrets in code**: Constitution enforces security practices

## Performance Considerations

- **Worker Pool**: Configurable workers (default 10)
- **Rate Limiting**: Per-CIDR leaky bucket
- **Memory**: Streaming CSV processing (low memory footprint)
