## Guidelines Index

**All guidelines in this file and the linked docs are mandatory, not recommendations.** Agents MUST read the relevant guideline doc before making changes in that area and MUST follow every rule.

| Layer | File | Scope |
|-------|------|-------|
| Consumer Core | [consumer/GUIDELINES.md](consumer/GUIDELINES.md) | Kafka consumption, offset management, retry, shutdown coordination |
| Transforms | [consumer/transforms/GUIDELINES.md](consumer/transforms/GUIDELINES.md) | CDC message transforms, service provider onboarding |
| Client | [internal/client/GUIDELINES.md](internal/client/GUIDELINES.md) | gRPC client, TLS/OIDC auth, SDK usage |
| Configuration | [internal/config/GUIDELINES.md](internal/config/GUIDELINES.md) | Options/Config/CompletedConfig pattern, Clowder, Viper |
| Metrics | [metrics/GUIDELINES.md](metrics/GUIDELINES.md) | OpenTelemetry metrics, Prometheus, Grafana alignment |

## Repository Context

The Kessel Inventory Consumer (KIC) is a standalone Kafka consumer group that subscribes to service-provider-owned Kafka topics and replicates resource updates to the Kessel Inventory API via gRPC. The Go module is `github.com/project-kessel/inventory-consumer`. It currently serves one service provider (Host Based Inventory / HBI) but is designed to onboard additional providers.

Key external dependencies:
- `confluent-kafka-go/v2` -- Kafka consumer (wraps librdkafka)
- `kessel-sdk-go` -- gRPC SDK for the Kessel Inventory API (v1beta2)
- `go-kratos/kratos/v2` -- logging framework
- `spf13/cobra` + `spf13/viper` + `spf13/pflag` -- CLI, config file, and flag management
- `opentelemetry` + `prometheus` -- metrics pipeline
- `redhatinsights/app-common-go` -- ClowdApp config injection for Red Hat cloud environments

## Architecture Overview

The processing pipeline flows through these stages:

```text
Kafka Topic -> Poll Loop (consumer.go) -> ParseHeaders -> ProcessMessage
  -> [ReportResource|DeleteResource]: ParseCreateOrUpdateMessage / ParseDeleteMessage -> gRPC Client
  -> [migration]: transforms package -> gRPC Client
```

Package boundaries:

| Package | Responsibility |
|---------|---------------|
| `cmd/` | Cobra commands (`start`, `readyz`), config wiring, signal handling |
| `consumer/` | Kafka consumer loop, message parsing, retry, offset management, shutdown |
| `consumer/auth/` | Kafka SASL authentication config |
| `consumer/retry/` | Retry parameter config |
| `consumer/transforms/` | CDC-to-protobuf transforms for raw Debezium messages |
| `consumer/types/` | Domain types for CDC source systems |
| `internal/client/` | gRPC client wrapping kessel-sdk-go |
| `internal/config/` | Top-level `OptionsConfig` aggregation, ClowdApp injection |
| `internal/common/` | Shared utilities (logging, reflection helpers) |
| `internal/mocks/` | testify/mock implementations of `Consumer` and `ClientProvider` |
| `metrics/` | OTel meter setup, Prometheus exporter, stats collection |

## Options / Config / CompletedConfig Pattern

Every configurable component uses a three-tier lifecycle: `Options` -> `Config` -> `CompletedConfig`. This is the central design pattern of the codebase. The full specification is in [internal/config/GUIDELINES.md](internal/config/GUIDELINES.md). Key points for cross-package awareness:

- `CompletedConfig` wraps an unexported `completedConfig` to prevent construction without calling `Complete()`.
- The required call sequence is: `options.Complete()` -> `options.Validate()` -> `NewConfig(options).Complete()`. See `cmd/start.go` for the canonical example.
- Every `Options` field must have a `mapstructure` tag, a corresponding pflag in `AddFlags()`, and a matching YAML key in `.inventory-consumer.yaml`.
- Adding a new component requires wiring into `internal/config/OptionsConfig`, the relevant command in `cmd/`, and the config file.

## Error Handling

- `Validate()` methods return `[]error` (a slice), not a single error. Collect all failures rather than short-circuiting.
- gRPC errors from `ClientProvider` methods are wrapped with `fmt.Errorf("failed to <action>: %w", err)`.
- Sentinel errors (`ErrClosed`, `ErrMaxRetries`, `ErrValidation`) drive control flow in the consumer loop. `ErrClosed` triggers consumer-level restart; `ErrMaxRetries` propagates fatal failure; `ErrValidation` (from header parsing) exits the consume loop.
- Domain-specific gRPC status codes (e.g., `codes.NotFound` on delete) are handled via custom error handler functions passed to `Retry()`.

## Testing

- **Framework**: `github.com/stretchr/testify` (assert + mock).
- **Pattern**: Table-driven tests with `t.Run(test.name, ...)` subtests.
- **Test scaffolding**: Use the `TestCase` struct and `TestSetup()` in the consumer package to bootstrap tests with Options, Config, CompletedConfig, MockConsumer, and InventoryConsumer.
- **Mocks**: `internal/mocks/mocks.go` provides `MockConsumer` (implements `consumer.Consumer` interface) and `MockClient` (implements `kessel.ClientProvider`). Both use testify/mock.
- **Flag coverage**: Every Options package has a test using `common.AllOptionsHaveFlags()` to verify that all `mapstructure`-tagged fields have registered pflags. Adding a field without a flag will fail this test.
- **Metrics coverage**: The `metricscollector_test.go` test uses reflection to verify every field on `MetricsCollector` is initialized after `New()`. Adding a metric field without registering it will fail this test.
- **Race tests**: `race_condition_test.go` tests concurrent shutdown/rebalance scenarios. Always run tests with the `-race` flag.
- **Test data**: Test message JSON and UUIDs are defined as package-level `const` values. Reuse existing constants rather than creating new fixtures.

## Build, Test, and Lint

```shell
make build          # Compile to ./bin/inventory-consumer
make test           # Run all tests with -race and coverage
make lint           # Run golangci-lint v2.6.2 via container
make lint-fix       # Run golangci-lint with --fix
```

Tests run with `-count=1 -race -short -covermode=atomic`. CI (GitHub Actions) runs both `build-test` and `golangci-lint` on every push/PR to main. The Go version is read from `go.mod`.

## Local Development

Local development requires the Kessel network of services. The recommended approach uses Podman Compose:

1. Clone and start Inventory API: `make inventory-up-relations-ready` (in the inventory-api repo)
2. Clone and start Relations API: `make relations-api-up` (in the relations-api repo)
3. Start KIC and dependencies: `make inventory-consumer-up` (in this repo)

This spins up KIC (3 replicas), a test HBI PostgreSQL database, Kafka topic creation, and a Debezium Kafka Connect cluster. The compose setup uses the external `kessel` Docker network. Config is read from `development/configs/full-setup.yaml`.

Other useful Makefile targets:
- `make setup-hbi-db` -- load the HBI schema into the test database
- `make setup-connectors` / `make delete-connectors` -- manage Debezium connectors
- `make check-connector-status` -- check connector health

## Docker and FIPS

The Dockerfile uses a two-stage build with Red Hat FIPS-validated base images. The runtime stage sets `GODEBUG=fips140=on`. The entrypoint is `inventory-consumer` with default command `start`. The binary runs as non-root user 1001.

## Deployment

KIC deploys as a ClowdApp on OpenShift (see `deploy/kessel-inventory-consumer.yaml`). Key deployment details:
- Default 3 replicas with a PodDisruptionBudget (minAvailable: 1)
- Readiness probe runs `inventory-consumer readyz` (gRPC health check against inventory-api)
- Config is injected via a Kubernetes Secret mounted at `/inventory/kic-config.yaml`
- ClowdApp integration: when `CLOWDER_ENABLED=true`, Kafka broker addresses and SASL credentials are injected from the ClowdApp config at startup (see `internal/config/config.go`)
- Tekton pipelines in `.tekton/` handle PR and push builds via Konflux

## CI/CD Pipelines

- **GitHub Actions**: `build-test.yml` (build + unit tests), `golangci-lint.yml` (linting), `codeql.yml` (security analysis), `jira-check.yml` (PR Jira ticket validation)
- **Tekton/Konflux**: PR and push pipelines build container images and push to Quay

## Monitoring

Prometheus metrics are served on port 9000 at `/metrics`. The Grafana dashboard ConfigMap in `dashboards/` queries specific metric names -- renaming or adding metrics requires updating the dashboard to match. See [metrics/GUIDELINES.md](metrics/GUIDELINES.md) for the full naming and registration rules.

## Adding a New Service Provider

This is the most common extension task. The full procedure is in [consumer/transforms/GUIDELINES.md](consumer/transforms/GUIDELINES.md). Summary of cross-package touchpoints:

1. Create domain types in `consumer/types/<provider>_types.go`
2. Create transform functions in `consumer/transforms/<provider>.go`
3. Add routing logic in `consumer/consumer.go` `ProcessMessage()` (new case or topic-based dispatch)
4. Add the topic to `.inventory-consumer.yaml` and `development/configs/full-setup.yaml`
5. Add the topic to the ClowdApp manifest in `deploy/`
6. Update mocks and tests as needed

## Adding a New API Operation

1. Add the method to `ClientProvider` interface in `internal/client/client.go`
2. Implement it on `KesselClient`
3. Add the method to `MockClient` in `internal/mocks/mocks.go`
4. Add the operation constant and `ProcessMessage` case in `consumer/consumer.go`
5. Add to `validOperations` map in `consumer/consumer.go`

## Naming Conventions and Guidelines

- **Go packages**: lowercase, single-word where possible
- **CLI flags**: dot-separated hierarchy with kebab-case segments (e.g., `consumer.auth.sasl-mechanism`)
- **YAML config keys**: match the `mapstructure` tag values exactly (kebab-case)
- **Metric names**: `consumer_stats_` prefix for Kafka stats, `consumer_` prefix for app-level counters, snake_case after prefix
- **Constants**: provider-specific constants follow `<Provider>ResourceType`, `<Provider>ReporterType` naming in `consumer/types/`
- **Operation types**: string constants in `consumer.go` (e.g., `OperationTypeReportResource`)
- **Go documentation comments**: Add Go doc comments to all newly added exported types and functions. Add comments to non-exported functions when their logic is not obvious. When unsure whether a comment is needed or how it should be written, follow the [Go doc comments guide](https://go.dev/doc/comment).

## Logging

Logging uses `go-kratos/kratos/v2/log`. The log level is configurable via the `log.level` YAML key (debug, info, warn, error, fatal; defaults to info). Each subsystem creates a `log.Helper` scoped with a `subsystem` key. Debug-level logging includes full config dumps (via `config.LogConfigurationInfo`) and full Kafka message payloads (key and value are logged via `msg.Value` and `msg.Key` in `Consume()` and `ProcessMessage()`). Never log secrets (passwords, tokens). When adding debug logging, limit output to sanitized metadata -- avoid logging raw message bodies in new code paths.
