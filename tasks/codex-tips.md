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

## When you can use Codex for review
- security review
- refactor suggestion
- performance review
- test gap analysis
- dead code detection

----
## Check codex code
After generating code by codex, you can check by these steps in the following: 
  1. Check the format
```bash
gofmt -w internal/otp/*.go
```
  2. Test implemented package
```bash
go test ./internal/otp -v
```
  3. Test percentage of coverage
```bash
go test ./internal/otp -cover
go test ./internal/otp -coverprofile=coverage.out
go tool cover -func=coverage.out
```
  4. Check by vet
```bash
go vet ./internal/otp
```
  5. Test whole of project
```bash
go test ./...
```
  6. Check diff
```bash
git add -N internal/otp/*.go
git diff --stat -- internal/otp
git diff -- internal/otp
```
