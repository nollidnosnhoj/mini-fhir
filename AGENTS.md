# AGENTS.md

This guide is for agentic coding tools working in this repository.

## Project overview
- Go-based in-memory FHIR DSTU3 server for CI/testing.
- Echo HTTP server, JSON-only.
- Validation reads StructureDefinition profiles and caches rules.
- Supported resources: Patient, Practitioner, PractitionerRole, Organization, Observation, Flag, Consent, AdvanceDirective, Location, Task.

## Build, run, test

### Build
- Build binary: `go build -o mini-fhir ./cmd/server`
- Run server: `./mini-fhir --addr :8080`
- Build Docker image: `docker build -t mini-fhir .`

### CLI flags
- `--addr`: HTTP listen address (default `:8080`).
- `--fhir-version`: FHIR version (default `dstu3`).
- `--seed`: Seed data glob pattern.
- `--seed-strict`: Fail on seed validation errors (default `true`).
- `--profile-cache`: Directory for StructureDefinition cache (default `.fhir-cache`).
- `--profile-cache-ttl`: Cache TTL for StructureDefinitions (default `24h`).
- `--profile-cache-version`: Cache version for StructureDefinitions (default `1`).

### Test
- Run all tests: `go test ./...`
- Run package tests: `go test ./internal/validation`
- Run a single test by name: `go test ./internal/validation -run TestMissingRequiredChoice`
- Run a single test file: `go test ./internal/validation -run Test -count=1`

### Lint/format
- Format Go files: `gofmt -w <file-or-dir>`
- No dedicated linter configured; use `go vet ./...` if needed.

## Code style guidelines

### Formatting
- Use `gofmt` on all Go files.
- Keep line lengths reasonable; use multi-line formatting for structs and composite literals.
- Prefer explicit field names in composite literals for complex structs.

### Imports
- Group imports into: standard library, blank line, third-party, blank line, local modules.
- Avoid unused imports; `go test` will fail on them.

### Naming
- Use Go naming conventions: CamelCase for exported types/functions, camelCase for unexported.
- Keep resource types and FHIR paths as DSTU3 canonical names (e.g., "Observation", "effectiveDateTime").
- File names should be short and descriptive (e.g., `store.go`, `required.go`).

### Types and structs
- Prefer strong types (structs/interfaces) over `map[string]any` unless required for generic FHIR JSON.
- For JSON structs, use `omitempty` on optional fields.
- For FHIR resources, embed `ResourceBase` and expose `GetResourceType()` via the interface.

### JSON handling
- Use strict JSON decoding with `json.Decoder.DisallowUnknownFields()` when handling API requests.
- Use `json.RawMessage` only when deferring full decode.
- Ensure `resourceType` is present and validated before processing.

### Error handling
- Return rich errors to API callers using `OperationOutcome`.
- For server internals, return `error` and let handlers map to HTTP status + OperationOutcome.
- Avoid panics except for truly unrecoverable initialization failures.

### Validation
- Validation is rule-based and loaded from DSTU3 StructureDefinitions.
- When adding new validation rules, update the cache version (`validation.CacheVersion`).
- Required fields are validated via `RequiredPaths` and `[x]` choice rules.

### Search
- `_sort=-date` for Observation should consider `effectiveDateTime`, `effectivePeriod.start`, then `issued`.
- `_include:iterate` is depth-limited (default 2).
- Keep search ordering deterministic for CI.

### Store
- Store operations are concurrency-safe and deterministic.
- Ensure `versionId` and `lastUpdated` are set on create/update.
- Clone resources when returning from the store to avoid mutation.

### API handlers
- Keep handlers thin: decode, validate, call store/search, marshal response.
- Use Bundle for search/history responses.
- For batch handling, respond with bundle entries and status per entry.

### Testing
- Prefer table-driven tests for validation and search.
- Use `httptest` for API handlers.
- Avoid network calls in tests; use local `ProfileStore` with test data.

## Repo rules
- No Cursor or Copilot rules detected in this repository.
- If new rules are added later, mirror them in this file.
