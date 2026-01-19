# Mini FHIR Server Plan (DSTU3)

## Scope + Constraints
- DSTU3 only.
- Resources: Patient, Practitioner, PractitionerRole, Organization, Observation, Flag, Consent, AdvanceDirective, Location, Task.
- JSON only.
- Search params: `_include`, `_include:iterate` (max depth 2), `_profile`, `_count`, `_sort`.
- `_sort=-date` for Observation uses `effective[x]` (dateTime -> period.start), fallback to `issued`.
- Validation: strict StructureDefinition; unknown elements rejected; optional profile validation when provided.

## Phase 1: Project Skeleton + Config
- Create module layout: `cmd/`, `internal/`.
- CLI flags: `--fhir-version` (dstu3), `--seed` (glob), optional `--seed-strict` (default true).
- Basic server bootstrap with router + JSON middleware.

## Phase 2: DSTU3 Model Foundation
- Core datatypes:
  - `Resource`, `Meta`, `Narrative`, `Reference`, `Identifier`, `CodeableConcept`, `Coding`,
    `HumanName`, `Address`, `Period`, `Quantity`, `Annotation`, `ContactPoint`, `Extension`,
    `Signature`, `Timing`, `Range`, `Age`, `Duration` (as needed).
- Strict JSON decoding with unknown-field rejection.
- Resource structs:
  - Patient, Practitioner, PractitionerRole, Organization, Observation, Flag, Consent,
    AdvanceDirective, Location.
- Resource registry for type discrimination.

## Phase 3: Validation Engine
- Load DSTU3 StructureDefinitions for supported resources + required datatypes.
- Extract validation rules:
  - required elements, cardinality, `[x]` choice types, primitive constraints.
- Validation pipeline:
  - JSON strict decode
  - structural + cardinality validation
  - reference type checks
  - optional profile validation
- Return `OperationOutcome` on errors.

## Phase 4: In-Memory Store + History
- Map: `ResourceType/id -> ResourceEntry`.
- Track `versionId`, `lastUpdated`, history list.
- Deterministic iteration ordering for CI.

## Phase 5: REST Endpoints
- CRUD + history:
  - `POST /{type}`, `GET /{type}/{id}`, `PUT /{type}/{id}`, `DELETE /{type}/{id}`
  - `GET /{type}/{id}/_history`, `GET /_history`
- `GET /metadata` CapabilityStatement.

## Phase 6: Search Engine
- `_profile`: filter by `meta.profile`.
- `_count`: page size.
- `_sort`:
  - Observation `_sort=-date`: `effective[x]` -> `issued`.
  - Generic `_sort=id`, `_sort=_lastUpdated`.
- `_include` / `_include:iterate`:
  - follow all references dynamically
  - max depth 2, cycle detection
  - include entries with `search.mode=include`.

## Phase 7: Batch/Transaction
- Batch: independent requests.
- Transaction: stage changes + rollback on failure.
- Bundle aggregation.

## Phase 8: `$validate`
- `POST /{type}/$validate` and `POST /$validate`.
- Validate against StructureDefinition + optional profile.

## Phase 9: Seed Loading
- `--seed` glob at startup.
- Accept Bundle or resource JSON.
- Validate on load; fail fast (strict).

## Phase 10: Tests
- CRUD + history tests.
- Validation + profile tests.
- Search + include tests.
- `_sort=-date` for Observation.
- Transaction rollback.
- Deterministic ordering for CI.
