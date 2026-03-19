# TESTING.md - Testing Practices

## Test Framework

- **Framework**: Go's `testing` package
- **Assertions**: Built-in testing functions

## Test Types

### Unit Tests

- **Location**: `*_test.go` alongside source files
- **Coverage**: 85% minimum
- **Execution**: `go test ./...`

### Integration Tests

- **Location**: `tests/integration/`
- **Purpose**: Test pipeline boundaries
- **Fixtures**: `tests/integration/testdata/`

### E2E Tests

- **Location**: `e2e/`
- **Environment**: Docker Compose (isolated networks)
- **Execution**: `bash e2e/run_e2e.sh`

## Quality Gates

| Gate | Command | Requirement |
|------|---------|-------------|
| Unit Tests | `go test ./...` | All pass |
| Coverage | `bash scripts/coverage_gate.sh` | ≥85% |
| E2E | `bash e2e/run_e2e.sh` | All pass |

## Mocking

- **Pattern**: Interface-based for testability
- **Examples**: 
  - `scanner.DialFunc` injectable for testing
  - `PressureFetcher` interface for mock implementations

## Test Structure

```go
func TestFeature_Name_Scenario(t *testing.T) {
    // Arrange
    // Act
    // Assert
}
```
