# TDD Evidence Contract

## Required Red Evidence

Every refactor increment must record:

- the exact failing test command run before the production change
- the failing test name or names that justify the increment
- the reason the failure is the intended failure, not a typo or environment issue

Acceptable examples:

- `go test ./cmd/port-scan -run TestRunMain_ScanWritesCSV -count=1`
- `go test ./pkg/scanapp -run TestRun_WhenResumeStateFileProvided_ContinuesFromNextIndex -count=1`

## Required Green Evidence

After the minimal code change, every increment must record:

- the exact command that proves the targeted test now passes
- the follow-up command that proves the surrounding package or affected suite still passes

Minimum standard:

- one targeted green command
- one broader regression command for the touched area

## Required Refactor Evidence

After green, the increment must record:

- what structural cleanup happened
- why the cleanup reduced coupling or clarified responsibility
- the command rerun after cleanup to prove behavior stayed green

This evidence belongs in `specs/001-solid-refactor-tdd/quickstart.md` under the relevant story.

## Invalid Shortcuts

The following do not count as TDD compliance:

- writing production code before the first failing test
- claiming a test would fail without running it
- using stale test output from a previous turn or previous commit
- changing both behavior and structure without explicitly separating the evidence
- relying on manual reasoning instead of an automated failing test

## Evidence Scope by Story

### User Story 1

- protect CLI composition and runtime-building behavior before extraction

### User Story 2

- protect dispatch, pause, and pressure coordination before seam extraction

### User Story 3

- protect result aggregation, progress reporting, and resume persistence before extraction

## Reviewer Checklist

For each increment, a reviewer must be able to answer yes to all of these:

- Was a failing test run first?
- Did the failing test correspond to the contract being protected?
- Was the code change minimal enough to explain clearly?
- Was the refactor step verified after cleanup?
- Does the evidence distinguish contract protection from structural movement?
