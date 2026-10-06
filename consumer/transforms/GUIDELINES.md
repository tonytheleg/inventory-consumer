## Architecture

This package converts Debezium CDC (Change Data Capture) messages from source systems into Kessel Inventory API v1beta2 protobuf request types. Each source system (called a "service provider") gets its own file containing transform functions. The consumer loop in `consumer.go` calls these transforms during `OperationTypeMigration` processing; standard `ReportResource`/`DeleteResource` operations use the generic parsers in `parsers.go` instead.

Transforms are the extension point for onboarding new CDC source systems that do not emit pre-formatted Kessel API payloads.

## Adding a New Service Provider

1. **Create a domain type file** in `consumer/types/` (e.g., `edge_types.go`). Define:
   - A message struct matching the Debezium CDC envelope (`Schema` + `Payload`).
   - A payload struct with fields matching the source database columns.
   - Constants for `ResourceType`, `ReporterType`, `ReporterInstanceID`, `APIHref`, `ConsoleHref`, and `ReporterVersion`. Follow the naming convention `<Provider>ResourceType`, `<Provider>ReporterType`, etc.
   - Custom `UnmarshalJSON` implementations for any field that may arrive as either a JSON array or a JSON-encoded string (see `GroupSlice` for the pattern).

2. **Create a transform file** in `consumer/transforms/` (e.g., `edge.go`). Implement exactly three exported functions following this naming pattern:
   - `Transform<Provider>ToReportResourceRequest(msg []byte) (*v1beta2.ReportResourceRequest, error)` -- for create/update/migration upserts.
   - `Transform<Provider>ToDeleteResourceRequest(msgValue []byte, msgKey []byte) (*v1beta2.DeleteResourceRequest, error)` -- for tombstone deletes.
   - `Is<Provider>Deleted(msgValue []byte) (bool, error)` -- to detect tombstone messages.

3. **Wire into the consumer** by adding a new case or routing logic in `ProcessMessage` (in `consumer.go`), gated on topic name or a header value that identifies the provider.

## Transform Function Contracts

### ReportResourceRequest Transform

- Accept raw `[]byte` message value (the Kafka message value).
- Unmarshal into the provider's domain type from `types/`.
- Build an intermediate `map[string]interface{}` payload, then marshal/unmarshal through JSON into `v1beta2.ReportResourceRequest`. This round-trip is required because `Representations.Common` and `Representations.Reporter` are `*structpb.Struct` fields that only populate correctly via JSON deserialization.
- The intermediate map must contain these top-level keys mapping to `ReportResourceRequest` fields:
  - `"type"` -> resource type constant
  - `"reporter_type"` -> reporter type constant
  - `"reporter_instance_id"` -> reporter instance ID constant
  - `"representations"` -> nested map with `"metadata"`, `"reporter"`, and `"common"` sub-maps
- `metadata` sub-map must include `local_resource_id` (the source system's primary key), `api_href`, `console_href`, and `reporter_version`.
- `reporter` sub-map holds provider-specific fields (e.g., `satellite_id`, `insights_id`).
- `common` sub-map holds cross-reporter fields (e.g., `workspace_id`).

### DeleteResourceRequest Transform

- Accept both `msgValue []byte` and `msgKey []byte`. Tombstone messages have empty/null values, so the resource ID must be extracted from the message key.
- Validate that the key is non-empty before attempting to unmarshal.
- Return a `DeleteResourceRequest` with a fully populated `ResourceReference` containing `ResourceType`, `ResourceId`, and a `ReporterReference` with `Type` set.
- Use an inline anonymous struct to unmarshal the key rather than importing a full domain type.

### Deletion Detection (Is<Provider>Deleted)

- Return `true` when `msgValue` is nil, empty, or the JSON literal `null`.
- Use `isEmptyJSON` (unexported helper) for the null/whitespace check.
- Always return `(bool, error)` even if the current implementation never returns an error -- the signature allows future implementations to inspect the value before deciding.

## Null and Empty Value Handling

- Use pointer types (`*string`) in domain structs for CDC columns that can be database NULL. JSON `null` deserializes to a nil pointer and flows through to `structpb.Struct` as a null value -- do not convert nil pointers to empty strings.
- The `groups` field uses the custom `GroupSlice` type because Debezium may emit it as either a JSON array or a stringified JSON array. Any similar polymorphic field in a new provider must use a custom type with `UnmarshalJSON`.
- **Collection access**: All transform implementations must validate collection length before accessing elements and return an error for empty/nil collections, rather than relying on panic recovery. The consumer's `safeProcessMessage` wrapper recovers unexpected panics and returns a non-nil error (triggering the consumer's retry path), but transforms should handle known cases explicitly with descriptive errors returned from `ProcessMessage`.

## Constants Conventions

- Resource type constants are lowercase identifiers agreed upon with the Kessel Inventory team (e.g., `"host"`).
- Reporter type constants are short abbreviations of the source system (e.g., `"hbi"` for Host Based Inventory).
- Reporter instance ID is a fixed string identifying the deployment (e.g., `"redhat"`).
- `APIHref` and `ConsoleHref` are placeholder URLs that may vary per environment in the future -- define them as constants now for consistency.
- All constants live in the provider's `types/` file, not in the transforms package.

## Relationship to parsers.go

`parsers.go` handles messages that arrive already formatted as Kessel API request payloads (with `operation` and `version` headers). Transforms handle raw CDC messages that need structural conversion. The distinction:

| Message source | Header `operation` | Processing path |
|---|---|---|
| Pre-formatted Kessel payload | `ReportResource` | `ParseCreateOrUpdateMessage` -> `v1beta2.ReportResourceRequest` |
| Pre-formatted Kessel payload | `DeleteResource` | `ParseDeleteMessage` -> `v1beta2.DeleteResourceRequest` |
| Raw CDC from source DB | `migration` | `transforms.IsHostDeleted` -> `TransformHostToReportResourceRequest` or `TransformHostToDeleteResourceRequest` |

New providers consuming raw CDC will use the `migration` path (or a new operation type) and call their transform functions.

## Testing Patterns

### Structure
- Use table-driven tests with the struct pattern: `name`, `message` (input bytes), `expectError`, `errorContains`, and `validate` (a function asserting on the output request).
- Define test message JSON as package-level `const` strings with interpolated test UUID constants.
- Define test UUIDs and error message substrings as package-level constants for reuse.

### Required Test Cases for ReportResourceRequest Transforms
1. Valid message with all fields populated -- verify every field on the output request.
2. Message with null optional fields -- verify nil pointer fields propagate as null in the `structpb.Struct` maps (use `.AsMap()` and `assert.Nil`).
3. Message with multiple items in a collection field -- verify the correct element is selected.
4. Invalid JSON input -- verify error contains the unmarshal error prefix.
5. Empty/missing required collections -- transforms must return a descriptive error for empty/nil collections. Tests assert the error message directly (no `defer/recover` needed).
6. Nil and empty byte slice inputs -- verify error returns.

### Required Test Cases for DeleteResourceRequest Transforms
1. Valid tombstone with a well-formed key -- verify `ResourceReference` fields.
2. Empty key, nil key -- verify specific error messages.
3. Invalid JSON in key -- verify error.
4. Key with missing or empty ID field -- verify error.

### Required Test Cases for Deletion Detection
1. Empty byte slice, nil, `"null"` string -- all return `true`.
2. Valid JSON, arbitrary non-empty bytes -- return `false`.

### Assertions
- Use `github.com/stretchr/testify/assert` for all assertions.
- Access `structpb.Struct` field values via `.AsMap()` and assert with `assert.Equal` or `assert.Nil`.
- For optional protobuf fields (pointer types like `*string`), dereference with `*field` in assertions.
