# CONVENTIONS.md - Coding Conventions

## Code Style

- **Standard**: Go standard conventions
- **Formatter**: `gofmt`
- **Linter**: golint, go vet

## SOLID Principles

From `.specify/memory/constitution.md`:

1. **Single Responsibility**: Each package/type/function has one clear responsibility
2. **Interface Segregation**: Minimal, purpose-specific interfaces
3. **Dependency Inversion**: High-level workflows depend on narrow abstractions

## Package Organization

- **Library-first**: Reusable logic in `pkg/`, CLI in `cmd/`
- **No god interfaces**: Minimal interfaces owned by consumers
- **No cyclic dependencies**: Clear dependency direction

## Error Handling

- **Fail-fast**: Return errors immediately for invalid inputs
- **Context**: Include context in error messages
- **Types**: Use custom error types where appropriate

## Testing Conventions

- **TDD Required**: Tests before implementation
- **Coverage**: 85% minimum (enforced by coverage_gate.sh)
- **Test Files**: `*_test.go` alongside implementation

## Documentation

- **Go Doc**: Public APIs documented with doc comments
- **Spec Files**: Implementation specs in `docs/specs/`
- **Plans**: Feature plans in `docs/plans/`
