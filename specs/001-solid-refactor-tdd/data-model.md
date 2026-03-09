# Refactor Data Model

## Responsibility Boundary

Definition:

- a named area of code with one primary purpose and one primary reason to change

Core fields:

- `name`
- `primary_purpose`
- `primary_reason_to_change`
- `current_owner_file`
- `target_owner_file`
- `protected_contracts`

Examples in this feature:

- CLI Composition Root
- Input Loader
- Runtime Builder
- Task Dispatcher
- Result Aggregator
- Resume Manager

## Refactor Increment

Definition:

- the smallest independently reviewable unit of structural change

Core fields:

- `story`
- `protected_behavior`
- `failing_test_command`
- `target_files`
- `structural_change`
- `green_verification_command`
- `post_refactor_verification_command`

## Protected Contract

Definition:

- an operator-visible behavior or reviewer-visible rule that must remain stable during refactor

Core fields:

- `contract_name`
- `scope`
- `evidence_tests`
- `breaking_change_trigger`

Examples:

- `validate json output`
- `scan cancellation exit mapping`
- `progress event presence`
- `resume completion without duplicates`

## Verification Evidence

Definition:

- the recorded proof that an increment followed red/green/refactor and preserved the intended contract

Core fields:

- `red_command`
- `red_observation`
- `green_command`
- `green_observation`
- `refactor_note`
- `post_refactor_command`
- `post_refactor_observation`

## Runtime Collaborator

Definition:

- a focused unit extracted from `pkg/scanapp` to own one runtime responsibility

Core fields:

- `collaborator_name`
- `responsibility_boundary`
- `inputs`
- `outputs`
- `dependencies`

Examples planned for this feature:

- `input_loader`
- `runtime_builder`
- `task_dispatcher`
- `pressure_monitor`
- `result_aggregator`
- `resume_manager`
