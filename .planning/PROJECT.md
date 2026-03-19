# port-scan-mk3

## What This Is

TCP port scanner CLI tool for network discovery and security auditing. Scans IP addresses/ CIDRs against configurable ports to detect open services.

## Core Value

Reliable, configurable TCP port scanning with rate limiting and pressure-aware pause control.

## Requirements

### Validated

- ✓ TCP port scanning with configurable timeout — existing
- ✓ CIDR CSV input parsing — existing
- ✓ Rich CSV input with policy metadata — existing
- ✓ Rate limiting (leaky bucket) — existing
- ✓ Pressure API auto-pause — existing (SimplePressureFetcher)
- ✓ Resume capability — existing
- ✓ CSV output — existing

### Active

- [ ] AUTH-01: Add AuthenticatedPressureFetcher constructor
- [ ] AUTH-02: Wire AuthenticatedPressureFetcher in scanapp.RunOptions

### Out of Scope

- IPv6 scanning — IPv4 only for v1
- UDP scanning — TCP only
- Service detection beyond open/closed — simple port check only

## Context

Brownfield project. Existing codebase has:
- SimplePressureFetcher (unauthenticated)
- AuthenticatedPressureFetcher struct (needs constructor)
- RunOptions struct (needs pressure fetcher option)

## Constraints

- **Tech Stack**: Go 1.24.x, standard library + golang.org/x
- **No Breaking Changes**: Existing CLI contracts must remain stable

## Key Decisions

| Decision | Rationale | Outcome |
|---------|-----------|---------|
| Library-first design | Reusable packages in pkg/, CLI in cmd/ | ✓ Good |
| Pressure-based pause | Auto-pause when external pressure high | ✓ Good |

---

*Last updated: 2026-03-19 after initialization*
