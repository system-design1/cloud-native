# AI Workflow — OTP Service Project

This project uses an AI-assisted engineering workflow with ChatGPT and Codex.

The purpose of this workflow is to keep implementation incremental, reviewable, testable, and token-efficient.

## Roles

### ChatGPT Responsibilities

ChatGPT acts as:

- senior software architect
- senior Go backend reviewer
- implementation planner
- technical reviewer
- workflow orchestrator
- project state manager
- documentation/state summarizer

ChatGPT is responsible for:

- breaking features into small phases
- preparing Codex analysis prompts
- reviewing Codex analysis
- correcting weak plans
- preparing final implementation prompts
- defining test strategy
- reviewing test outputs
- deciding whether a slice is safe to commit
- managing roadmap and deferred concerns
- helping with thread reset and project memory

### Codex Responsibilities

Codex acts as:

- focused implementation assistant
- code modifier
- test writer
- repository-aware implementer

Codex is responsible for:

- analyzing the codebase when asked
- implementing only approved scope
- modifying only allowed files
- adding focused tests
- running gofmt/tests when requested
- summarizing changed files and test results

Codex should not own overall architecture direction.

## Core Workflow

Every meaningful implementation follows this cycle:

```text
1. ChatGPT prepares analysis prompt
2. Codex analyzes only, without coding
3. User brings Codex analysis back to ChatGPT
4. ChatGPT reviews analysis
5. ChatGPT explains strengths, weaknesses, tradeoffs
6. ChatGPT prepares final implementation prompt
7. Codex implements approved small scope
8. User runs tests/manual validation
9. User brings outputs back to ChatGPT
10. ChatGPT reviews outputs
11. ChatGPT confirms whether to commit
12. User commits
13. Move to next small phase
```

## Analysis Before Implementation

Codex analysis prompts must explicitly say:

```text
Do not modify files yet.
Do not implement anything yet.
Only analyze.
```

The analysis should include:

- recommended implementation approach
- exact files to change
- dependency and wiring impact
- risks/tradeoffs
- edge cases
- error behavior
- tests required
- manual validation
- deferred concerns
- whether the task should be split smaller

## Incremental Implementation Strategy

Implementation must be split into small slices.

Good slice examples:

```text
Only add domain models and interfaces.
Only add Redis adapter.
Only wire config.
Only add handler tests.
Only add repository migration.
```

Bad slice examples:

```text
Implement complete OTP flow and observability and rate limiting together.
Refactor all packages while adding a feature.
Add metrics, tracing, provider router, and circuit breaker in one step.
```

## Scope Management Rules

Each implementation prompt should include:

- exact files allowed to change
- exact behavior to add
- what must not be changed
- what is deferred
- tests to add
- commands to run

Codex should be constrained to the smallest useful diff.

## Testing Workflow

Each slice should include tests appropriate to its scope:

- service tests for domain/application logic
- handler tests for HTTP mapping
- repository integration-style tests for Redis/PostgreSQL
- config tests for env parsing and validation
- provider tests for SMS behavior

Common commands:

```bash
gofmt -w <changed-go-files>
go test -count=1 ./internal/otp -v
go test -count=1 ./internal/api -v
go test -count=1 ./internal/repository -v
go test -count=1 ./internal/config -v
go test -count=1 ./internal/sms -v
go test -count=1 ./...
```

Coverage checks can be used when useful:

```bash
go test -count=1 ./internal/otp -coverprofile=coverage.out
go tool cover -func=coverage.out
```

## Manual Validation Workflow

For runtime-sensitive features, manual tests are required.

Examples:

- call `/v1/otp/send`
- read Redis OTP state
- read fake SMS debug code
- call `/v1/otp/verify`
- query `otp_requests`
- query `otp_verifications`
- verify rate limiting behavior
- verify resend protection behavior

Manual validation should include:

- request command
- expected HTTP response
- database query if relevant
- Redis key checks if relevant
- cleanup commands if needed

## Commit Strategy

Commit after each stable slice.

Before commit:

```bash
go test -count=1 ./...
git status --short
git diff --stat
```

Commit messages should be small and focused.

Examples:

```text
Add Redis OTP store
Implement OTP verification service
Add OTP HTTP handlers
Wire OTP service into server
Add OTP verification logging repository
Add Redis OTP send rate limiter
Wire OTP send rate limiter config
```

Do not mix task docs with implementation commits unless intentionally desired.

## Thread Management Strategy

When the conversation becomes heavy:

1. generate/update project state docs
2. commit the docs
3. start a new thread
4. bootstrap the new thread with a concise summary

Useful docs:

```text
docs/current-state.md
docs/deferred-concerns.md
docs/architecture.md
docs/roadmap.md
docs/ai-workflow.md
```

New threads should include:

- current state
- completed phases
- active goal
- deferred concerns
- workflow rules
- constraints

## Token Optimization Strategy

To reduce token usage:

- keep Codex prompts scoped
- avoid asking Codex to read unrelated files
- avoid large all-in-one tasks
- summarize after major phases
- commit frequently
- use docs/current-state.md for thread reset
- avoid repeating full history
- use focused file lists in implementation prompts

## Context Reset Strategy

Before starting a new thread:

Ask ChatGPT to generate or update:

- current-state.md
- deferred-concerns.md
- architecture.md
- roadmap.md
- ai-workflow.md

Then start the new thread with:

```text
We are continuing an existing project.
Here is the current state:
...
Current goal:
...
Important constraints:
...
Use the same incremental ChatGPT + Codex workflow.
```

## Implementation Constraints

Important constraints used throughout this project:

- keep diffs small
- avoid broad refactors
- prefer explicit/simple code
- preserve existing architecture style
- use idiomatic Go
- keep handlers thin
- keep domain logic in service layer
- keep infrastructure details in repository/adapters
- use interfaces where they improve testability/boundaries
- avoid premature abstractions
- keep tests deterministic
- do not expose OTP codes in API/logs
- do not store plaintext OTP in main OTP state
- update env.example when new env/config is added

## Review Expectations

When reviewing Codex output:

Check:

- whether scope was respected
- whether unrelated files changed
- whether tests were added
- whether tests actually passed
- whether manual validation is needed
- whether error behavior is correct
- whether architecture boundaries were preserved
- whether new runtime/config values were documented
- whether implementation is safe to commit

## Current Workflow Status

This workflow has already been used successfully for the OTP service implementation, including:

- domain foundation
- Redis store
- SendOTP
- VerifyOTP
- HTTP handlers
- request logging
- verification logging
- fake SMS provider
- debug code capture
- resend protection
- send rate limiting
- config/env wiring
