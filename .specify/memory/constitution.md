# Port Scan MKIII Constitution
<!-- Example: Spec Constitution, TaskFlow Constitution, etc. -->

## Core Principles

### I. Library-First
<!-- Example: I. Library-First -->
- Every feature starts as a standalone library
- Libraries must be self-contained, independently testable, documented
- Clear purpose required - no organizational-only libraries

### II. CLI Interface
<!-- Example: II. CLI Interface -->
- Every library exposes functionality via CLI
- Text in/out protocol: stdin/args → stdout, errors → stderr
- Support JSON + human-readable formats

### III. Test-First (NON-NEGOTIABLE)
<!-- Example: III. Test-First (NON-NEGOTIABLE) -->
- TDD mandatory: Tests written → User approved → Tests fail → Then implement
- Red-Green-Refactor cycle strictly enforced

### IV. Integration Testing
<!-- Example: IV. Integration Testing -->
- Focus areas requiring integration tests:
  - New library contract tests
  - Contract changes
  - Inter-service communication
  - Shared schemas

### V. End-to-End Testing (e2e test)

- Focus on using docker compose to create a test area that:
  - has multiple mock server running, with opened TCP ports
  - has isolated network that not scanning behavior will afftec real machines and environments
  - the test area must capable of testing all functions and features that the Port Scan MKIII provieds
  - Add any necessary services into test areas to ensure e2e test covers all scenarios

### VI. Observability
<!-- Example: V. Observability, VI. Versioning & Breaking Changes, VII. Simplicity -->
- Text I/O ensures debuggability
- Structured logging required 

### VII. Versioning
<!-- Example: V. Observability, VI. Versioning & Breaking Changes, VII. Simplicity -->
- MAJOR.MINOR.BUILD format
- Every release must have release note includes:
  - New features
  - Bug fixes
  - Breaking changes (must have migration guideline)
<!-- Example: Text I/O ensures debuggability; Structured logging required; Or: MAJOR.MINOR.BUILD format; Or: Start simple, YAGNI principles -->

## Technology stack requirements
<!-- Example Section: Additional Constraints, Security Requirements, Performance Standards, etc. -->
<!-- Example content: Technology stack requirements, compliance standards, deployment policies, etc. -->
<!-- Example Section: Development Workflow, Review Process, Quality Gates, etc. -->
<!-- Example content: Code review requirements, testing gates, deployment approval process, etc. -->
- Golang as main language
  - use Golang built-in net package to implement TCP port scan
    - use dial to try to establish tcp connection

## Quality Gates
<!-- Example content: testing gates -->
- testing gates:
  - unit test coverages > 85%
  - e2e test must all passed
  - integration tests must all passed

## Governance
<!-- Example: Constitution supersedes all other practices; Amendments require documentation, approval, migration plan -->
- All PRs/reviews must verify compliance
- Complexity must be justified
- Amendments require documentation, approval, migration plan

**Version**: 1.0.0 | **Ratified**: 2026-03-01 | **Last Amended**: 2026-03-01
