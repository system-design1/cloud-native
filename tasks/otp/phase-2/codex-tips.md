# Codex tips 

## CLI vs Extension
- The situations that you may better use codex extension:
  - design iteration
  - small diffs
  - reviewing code
  - interactive development
  - reading files
  - controlled edits
  - low-risk changes
- The situations that you may better use codex cli:
  - autonomous tasks
  - repo-wide changes
  - long-running work
  - command execution loops
  - automation
  - large refactors
  - scripting workflows
  - generate integration tests
  - refactor repository package
  - benchmark scenarios
  - run/fix/test loops

## How to switch between CLI and extension

**Note**: Only one active agent task at a time.
- before switch between them:
  - you must complete the task
  - you must review the task
  - add a commit or checkpoint

- This is a safe workflow:
Extension
→ review
→ commit

CLI
→ review
→ commit

Extension
→ review
→ commit

- This is dangerous workflow:
CLI changing files
+
Extension changing same files simultaneously


------------