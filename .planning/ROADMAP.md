# Roadmap: Authenticated Pressure API Support

## Phase 1: Add Authenticated Pressure API CLI Flags

**Goal**: Add CLI flags for authenticated pressure API configuration

### Requirements

- [ ] **AUTH-01**: Add CLI flags for authenticated pressure API

### Success Criteria

1. New flags available:
   - `-pressure-auth-url` (auth endpoint URL)
   - `-pressure-auth-client-id` (OAuth client ID)
   - `-pressure-auth-client-secret` (OAuth client secret)

### Tasks

1. Add fields to `config.Config`:
   - `PressureAuthURL`
   - `PressureAuthClientID`
   - `PressureAuthClientSecret`

2. Add flag definitions in `config.Parse()`

3. Add validation (flags optional, no validation needed)

---

## Phase 2: Wire AuthenticatedPressureFetcher in Run

**Goal**: Use authenticated fetcher when auth flags provided

### Requirements

- [ ] **AUTH-02**: Wire AuthenticatedPressureFetcher in scanapp.Run

### Success Criteria

1. When auth flags provided → use AuthenticatedPressureFetcher
2. When only basic pressure-api → use SimplePressureFetcher

### Tasks

1. In `scanapp.Run()`, check if auth flags provided
2. If auth flags present, create AuthenticatedPressureFetcher
3. Otherwise, use SimplePressureFetcher (existing behavior)

---

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| AUTH-01 | Phase 1 | Pending |
| AUTH-02 | Phase 2 | Pending |

**Coverage:**
- v1 requirements: 2 total
- Mapped to phases: 2
- Unmapped: 0 ✓
