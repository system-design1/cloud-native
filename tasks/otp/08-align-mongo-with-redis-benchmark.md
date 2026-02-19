# 08 - Align Mongo Benchmark with Redis Benchmark (Exact Behavior Match)

## Goal

Modify Mongo benchmark implementation so that it behaves exactly like
Redis benchmark endpoints.

This task must strictly follow the deterministic Cursor style: - Atomic
change - Limited file scope - No refactoring outside defined files - No
behavior drift - No scope creep

------------------------------------------------------------------------

## TASK

Align Mongo GET/SET behavior to be identical to Redis benchmark
semantics.

IMPORTANT: - Do NOT modify Redis implementation. - Do NOT change
routes. - Only align Mongo implementation. - Modify only specified Mongo
files.

------------------------------------------------------------------------

## INPUT

Files to modify:

1)  internal/repository/mongo_benchmark_repo.go\
2)  internal/api/mongo_benchmark_handlers.go

------------------------------------------------------------------------

## REQUIRED CHANGES

### 1️⃣ Key Format --- Must Match Redis

Remove generic key usage.

Handlers must accept:

-   tenant (required)
-   phone (required)
-   value (for SET)
-   ttl (optional)

Key must be constructed exactly as:

    fmt.Sprintf("otp:%s:%s", tenant, phone)

Raw key parameter must not be accepted anymore.

------------------------------------------------------------------------

### 2️⃣ SET Behavior --- Match Redis Semantics

-   Default TTL = 120 seconds
-   If ttl invalid or \<= 0 → fallback to 120s
-   Upsert document with:
    -   \_id = formatted key
    -   value = provided value
    -   expires_at = time.Now().Add(ttl)

Must NOT create any index inside request path.

------------------------------------------------------------------------

### 3️⃣ GET Behavior — Simulate Redis TTL

When GET is called:

- If no document exists: return `{"found": false}`

- If document exists:
  - If `time.Now() > expires_at`:
    - Delete the document
    - Return `{"found": false}`
  - Else:
    - Return `{"found": true, "value": "<value>"}`


------------------------------------------------------------------------

### 4️⃣ Response Shape --- Must Be Identical to Redis

SET response: {"ok": true}

GET found: {"found": true, "value": "..."}

GET not found: {"found": false}

No additional fields allowed.

------------------------------------------------------------------------

### 5️⃣ Validation

If tenant or phone missing: - Use middleware.ErrorHandler - Return bad
request error

------------------------------------------------------------------------

## OUTPUT

Modify exactly these files:

-   internal/repository/mongo_benchmark_repo.go
-   internal/api/mongo_benchmark_handlers.go

Do NOT modify any other file.

------------------------------------------------------------------------

## RULES

-   No explanation
-   No analysis
-   No extra text
-   Output only modified files
-   Keep English comments only
-   Do not refactor unrelated code
-   Do not change route registration
-   Do not modify Redis code
