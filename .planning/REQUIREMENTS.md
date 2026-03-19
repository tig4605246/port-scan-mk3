# Requirements: port-scan-mk3

**Defined:** 2026-03-19
**Core Value:** Reliable, configurable TCP port scanning with rate limiting and pressure-aware pause control

## v1 Requirements

### Authentication

- [ ] **AUTH-01**: Add CLI flags for authenticated pressure API
- [ ] **AUTH-02**: Wire AuthenticatedPressureFetcher in scanapp.Run

## v2 Requirements

(None currently)

## Out of Scope

| Feature | Reason |
|---------|--------|
| IPv6 scanning | IPv4 only for v1 |
| UDP scanning | TCP only |
| Service detection | Simple port check only |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| AUTH-01 | Phase 1 | Pending |
| AUTH-02 | Phase 2 | Pending |

**Coverage:**
- v1 requirements: 2 total
- Mapped to phases: 2
- Unmapped: 0 ✓

---

*Requirements defined: 2026-03-19*
*Last updated: 2026-03-19 after initialization*
