# Evidence-Driven Delivery (Showboat + Rodney)

Use this workflow when you must deliver proof from real execution, not narrative claims.

## Step 0: Ask for required inputs

- Ask the user for:
  - `TASK_SLUG` (kebab-case)
  - verification command list (tests/build/key checks)
  - whether this is a Web task (`yes`/`no`)
  - if Web task: target URL and selectors/assertions to validate

If any required input is missing, pause and ask before running commands.

## Step 1: Environment checks (stop on failure)

<execute_command>
<command>command -v uvx >/dev/null && echo "uvx ready" || { echo "uvx not found"; exit 1; }</command>
</execute_command>

<execute_command>
<command>git status --short</command>
</execute_command>

If `uvx` is missing, stop and report installation requirement.

## Step 2: Create evidence document

Run:

```bash
mkdir -p docs/demos
uvx showboat --help
uvx showboat init "docs/demos/<TASK_SLUG>-demo.md" "<TASK_SLUG> evidence demo"
uvx showboat note "docs/demos/<TASK_SLUG>-demo.md" "Evidence workflow started for <TASK_SLUG>."
```

## Step 3: Record project verification commands

For each user-confirmed verification command:

```bash
uvx showboat note "docs/demos/<TASK_SLUG>-demo.md" "Verify: <purpose>"
uvx showboat exec "docs/demos/<TASK_SLUG>-demo.md" bash "<command>"
```

Rules:
- Run commands exactly as provided/confirmed.
- If a command fails, stop and report failure with evidence path.
- Do not hand-edit `docs/demos/<TASK_SLUG>-demo.md`.
- If you need to undo the latest mistaken entry, use:

```bash
uvx showboat pop "docs/demos/<TASK_SLUG>-demo.md"
```

## Step 4: Conditional Rodney flow (Web tasks only)

If Web task is `yes`, run this sequence and capture each critical step in Showboat.

```bash
uvx showboat note "docs/demos/<TASK_SLUG>-demo.md" "Start Rodney web validation."
uvx showboat exec "docs/demos/<TASK_SLUG>-demo.md" bash "uvx rodney --help >/dev/null"
uvx showboat exec "docs/demos/<TASK_SLUG>-demo.md" bash "uvx rodney start"
uvx showboat exec "docs/demos/<TASK_SLUG>-demo.md" bash "uvx rodney open \"<TARGET_URL>\""
uvx showboat exec "docs/demos/<TASK_SLUG>-demo.md" bash "uvx rodney waitload && uvx rodney waitidle"
uvx showboat exec "docs/demos/<TASK_SLUG>-demo.md" bash "uvx rodney exists \"<PRIMARY_SELECTOR>\""
uvx showboat exec "docs/demos/<TASK_SLUG>-demo.md" bash "uvx rodney visible \"<PRIMARY_SELECTOR>\""
uvx showboat exec "docs/demos/<TASK_SLUG>-demo.md" bash "uvx rodney assert \"<PRIMARY_ASSERT>\""
uvx showboat exec "docs/demos/<TASK_SLUG>-demo.md" bash "uvx rodney screenshot docs/demos/<TASK_SLUG>-web.png && echo docs/demos/<TASK_SLUG>-web.png"
uvx showboat image "docs/demos/<TASK_SLUG>-demo.md" "docs/demos/<TASK_SLUG>-web.png"
uvx showboat exec "docs/demos/<TASK_SLUG>-demo.md" bash "uvx rodney stop"
```

If any key Rodney step fails, stop immediately and report failure + next action.

## Step 5: Verify demo integrity (mandatory gate)

```bash
uvx showboat verify "docs/demos/<TASK_SLUG>-demo.md"
uvx showboat extract "docs/demos/<TASK_SLUG>-demo.md"
```

If `showboat verify` fails, stop and report diff results. Do not claim completion.

## Step 6: Delivery summary format

Report:
- demo file path: `docs/demos/<TASK_SLUG>-demo.md`
- key verification commands executed
- Web evidence files (if any)
- final verification status from `showboat verify`

Completion rule:
- If tests/checks/verify fail, explicitly report failed delivery.
- Never claim "done" without successful `showboat verify` evidence.
