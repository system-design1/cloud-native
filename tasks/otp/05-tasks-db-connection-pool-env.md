# DB Connection Pool via ENV (Postgres) — Tasks for Cursor

Goal:
- Read DB connection pool settings from ENV and apply them to `sql.DB` pool.
- Add variables to `.env` and `env.example`.
- Keep current behavior as defaults (25 open, 5 idle, 5m lifetime, 10m idle time).

---

## Prompt 01 — Extend DatabaseConfig + load from ENV

TASK: Extend database config to include pool settings loaded from ENV with defaults.
INPUT:
- File: internal/config/config.go
- Update `type DatabaseConfig` to include:
  - MaxOpenConns int
  - MaxIdleConns int
  - ConnMaxLifetime time.Duration
  - ConnMaxIdleTime time.Duration
- In `loadDatabaseConfig`, read these env vars (with defaults):
  - DB_MAX_OPEN_CONNS (default 25)
  - DB_MAX_IDLE_CONNS (default 5)
  - DB_CONN_MAX_LIFETIME (default 5m)
  - DB_CONN_MAX_IDLE_TIME (default 10m)
- Validate:
  - MaxOpenConns >= 1
  - MaxIdleConns >= 0
  - MaxIdleConns <= MaxOpenConns
  - Durations must parse (time.ParseDuration)
- Set these fields on `cfg.Database`
OUTPUT: Modify exactly one file: internal/config/config.go
RULES:
- No explanation
- No analysis
- No extra text
- Output only the result

---

## Prompt 02 — Apply pool settings in db.NewConnectionPool

TASK: Use DatabaseConfig pool settings when configuring sql.DB.
INPUT:
- File: internal/db/db.go
- Replace hard-coded pool config with cfg values:
  - db.SetMaxOpenConns(cfg.MaxOpenConns)
  - db.SetMaxIdleConns(cfg.MaxIdleConns)
  - db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
  - db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
- Keep everything else unchanged
OUTPUT: Modify exactly one file: internal/db/db.go
RULES:
- No explanation
- No analysis
- No extra text
- Output only the result

---

## Prompt 03 — Update config tests for new defaults

TASK: Update config tests to cover default DB pool values.
INPUT:
- File: internal/config/config_test.go
- Add assertions that defaults are:
  - MaxOpenConns == 25
  - MaxIdleConns == 5
  - ConnMaxLifetime == 5 * time.Minute
  - ConnMaxIdleTime == 10 * time.Minute
- Keep existing tests intact
OUTPUT: Modify exactly one file: internal/config/config_test.go
RULES:
- No explanation
- No analysis
- No extra text
- Output only the result

---

## Prompt 04 — Add ENV vars to .env

TASK: Add DB pool env vars to .env.
INPUT:
- File: .env
- Add these lines near other DB_* variables:
  - DB_MAX_OPEN_CONNS=25
  - DB_MAX_IDLE_CONNS=5
  - DB_CONN_MAX_LIFETIME=5m
  - DB_CONN_MAX_IDLE_TIME=10m
OUTPUT: Modify exactly one file: .env
RULES:
- No explanation
- No analysis
- No extra text
- Output only the result

---

## Prompt 05 — Add ENV vars to env.example

TASK: Add DB pool env vars to env.example.
INPUT:
- File: env.example
- Add these lines near other DB_* variables:
  - DB_MAX_OPEN_CONNS=25
  - DB_MAX_IDLE_CONNS=5
  - DB_CONN_MAX_LIFETIME=5m
  - DB_CONN_MAX_IDLE_TIME=10m
OUTPUT: Modify exactly one file: env.example
RULES:
- No explanation
- No analysis
- No extra text
- Output only the result
