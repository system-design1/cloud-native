# 07 - MongoDB Benchmark (GET/SET) Tasks for Cursor

## Goal

Add MongoDB to the project and implement benchmark endpoints for:

-   POST /v1/mongo/set
-   GET /v1/mongo/get

This must follow the exact deterministic style used in Task 6: - Atomic
prompts - Single-file modification per prompt - Strict output rules - No
scope creep

------------------------------------------------------------------------

## Prompt 01 --- Add MongoDB service to docker-compose

TASK: Add a MongoDB service for local development and connect it to
app-network.

INPUT: File: docker-compose.yml

Add service: - image: mongo:7 - container_name: go-backend-mongo -
ports: "27017:27017" - environment: - MONGO_INITDB_ROOT_USERNAME=root -
MONGO_INITDB_ROOT_PASSWORD=secret - healthcheck: - test: \["CMD",
"mongosh", "--quiet", "mongodb://root:secret@localhost:27017/admin",
"--eval", "db.runCommand({ ping: 1 }).ok"\] - interval: 10s - timeout:
5s - retries: 10 - networks: app-network

Ensure api depends_on includes mongo with condition service_healthy.

OUTPUT: Modify exactly one file: docker-compose.yml

RULES: - No explanation - No analysis - No extra text - Output only the
result

------------------------------------------------------------------------

## Prompt 02 --- Add Mongo ENV vars

TASK: Add MongoDB env vars to .env and env.example.

Add: MONGO_URI=mongodb://root:secret@mongo:27017/admin?authSource=admin
MONGO_DB=otp_bench MONGO_COLLECTION=benchmark_kv MONGO_MAX_POOL_SIZE=200
MONGO_MIN_POOL_SIZE=20 MONGO_CONNECT_TIMEOUT=2s
MONGO_SERVER_SELECTION_TIMEOUT=2s MONGO_SOCKET_TIMEOUT=2s
MONGO_HEARTBEAT_INTERVAL=10s

OUTPUT: Modify: - .env - env.example

RULES: - No explanation - No analysis - No extra text - Output only the
result

------------------------------------------------------------------------

## Prompt 03 --- Extend Config

TASK: Add MongoConfig struct and loader in internal/config/config.go.

Requirements: - Add Mongo MongoConfig to Config struct - Add MongoConfig
struct - Add loadMongoConfig function - Validate required fields - Parse
durations using time.ParseDuration - Set sensible defaults

OUTPUT: Modify exactly one file: internal/config/config.go

RULES: - No explanation - No analysis - No extra text - Output only the
result

------------------------------------------------------------------------

## Prompt 04 --- Create Mongo Client

TASK: Create Mongo client initializer.

File: internal/mongo/mongo.go

Requirements: - Use official mongo-driver - Apply pool settings - Apply
timeout settings - Connect with 2s context timeout - Ping with
readpref.Primary() - Return wrapped errors - English comments only

OUTPUT: Exactly one file: internal/mongo/mongo.go

RULES: - No explanation - No analysis - No extra text - Output only the
result

------------------------------------------------------------------------

## Prompt 05 --- Create Mongo Benchmark Repository

TASK: Create repository for benchmark SET/GET.

File: internal/repository/mongo_benchmark_repo.go

Requirements: - Use \_id as key - Store value string - Store expires_at
time.Time - SetBenchmarkKey with upsert - GetBenchmarkKey returning
value or mongo.ErrNoDocuments - No index creation inside request path -
Wrap errors properly - English comments only

OUTPUT: Exactly one file: internal/repository/mongo_benchmark_repo.go

RULES: - No explanation - No analysis - No extra text - Output only the
result

------------------------------------------------------------------------

## Prompt 06 --- Create API Handlers

TASK: Add Gin handlers for Mongo benchmark.

File: internal/api/mongo_benchmark_handlers.go

Endpoints: POST /v1/mongo/set GET /v1/mongo/get

Requirements: - key and value from query params - ttl optional (default
120s) - Use middleware.ErrorHandler - Return structured JSON - English
comments only

OUTPUT: Exactly one file: internal/api/mongo_benchmark_handlers.go

RULES: - No explanation - No analysis - No extra text - Output only the
result

------------------------------------------------------------------------

## Prompt 07 --- Wire Routes

TASK: Register routes under /v1/mongo.

File: internal/api/routes.go

Requirements: - Add route group /v1/mongo - POST /set - GET /get -
Update SetupRoutes signature - Keep existing routes unchanged

OUTPUT: Modify exactly one file: internal/api/routes.go

RULES: - No explanation - No analysis - No extra text - Output only the
result

------------------------------------------------------------------------

## Prompt 08 --- Wire in main.go

TASK: Initialize Mongo client and repository.

File: cmd/server/main.go

Requirements: - Create client using config.Mongo - Handle error
properly - Defer Disconnect - Create repository - Pass repo into
SetupRoutes - Keep everything else unchanged

OUTPUT: Modify exactly one file: cmd/server/main.go

RULES: - No explanation - No analysis - No extra text - Output only the
result

------------------------------------------------------------------------

## Prompt 09 --- Update go.mod

TASK: Add official MongoDB driver dependency.

File: go.mod

Requirement: - Add go.mongodb.org/mongo-driver - Reflect go mod tidy
changes

OUTPUT: Modify exactly one file: go.mod

RULES: - No explanation - No analysis - No extra text - Output only the
result
