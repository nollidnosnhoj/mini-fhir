# Mini FHIR (DSTU3)

Lightweight in-memory FHIR server for CI/testing. Supports DSTU3 resources: Patient, Practitioner, PractitionerRole, Organization, Observation, Flag, Consent, AdvanceDirective, Location, Task.

## Features
- In-memory store with history
- Search params: `_include`, `_include:iterate`, `_profile`, `_count`, `_sort`
- `_sort=-date` on Observation (effective[x] -> issued)
- `$validate` with StructureDefinition checks and optional profile
- Batch bundle handling
- Seed loading via CLI flag
- StructureDefinition cache with TTL/version

## Build and run

```bash
go build -o mini-fhir ./cmd/server
./mini-fhir --addr :8080
```

## CLI flags
- `--addr`: HTTP listen address (default `:8080`).
- `--fhir-version`: FHIR version (default `dstu3`).
- `--seed`: Seed data glob pattern.
- `--seed-strict`: Fail on seed validation errors (default `true`).
- `--profile-cache`: Directory for StructureDefinition cache (default `.fhir-cache`).
- `--profile-cache-ttl`: Cache TTL for StructureDefinitions (default `24h`).
- `--profile-cache-version`: Cache version for StructureDefinitions (default `1`).

## Tests

```bash
go test ./...
```

## Seed data

```bash
./mini-fhir --seed "./fixtures/*.json"
```

## Validation cache

```bash
./mini-fhir --profile-cache .fhir-cache --profile-cache-ttl 24h --profile-cache-version 1
```

## Docker

```bash
docker build -t mini-fhir .
docker run --rm -p 8080:8080 mini-fhir
```
