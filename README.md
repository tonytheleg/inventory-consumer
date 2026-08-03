# Kessel Inventory Consumer

The Kessel Inventory Consumer (KIC) is a standalone dedicated Kafka consumer group used to expose an eventing based entry point to the Kessel Inventory API. Its purpose is to subscribe to Service Provider owned Kafka topics and ensure reporter resource updates are replicated to Inventory API through events.

## Project Structure

```text
cmd/                       CLI commands (start, readyz) and config wiring
consumer/                  Core Kafka consumer loop, message parsing, retry, shutdown
  transforms/              CDC-to-protobuf transforms for Debezium messages
  types/                   Domain types for CDC source systems
internal/client/           gRPC client wrapping kessel-sdk-go
internal/config/           Top-level config aggregation, ClowdApp injection
metrics/                   OpenTelemetry metrics and Prometheus exporter
deploy/                    OpenShift ClowdApp manifests
development/               Local Podman Compose setup and configs
```

## Documentation

- [AGENTS.md](./AGENTS.md) -- AI agent onboarding, architecture overview, and conventions
- Guidelines -- mandatory rules for each subsystem:
  - [consumer/GUIDELINES.md](./consumer/GUIDELINES.md) -- Kafka consumption, offset management, retry, shutdown
  - [consumer/transforms/GUIDELINES.md](./consumer/transforms/GUIDELINES.md) -- CDC transforms, service provider onboarding
  - [internal/client/GUIDELINES.md](./internal/client/GUIDELINES.md) -- gRPC client, TLS/OIDC auth, SDK usage
  - [internal/config/GUIDELINES.md](./internal/config/GUIDELINES.md) -- Options/Config/CompletedConfig pattern, Clowder, Viper
  - [metrics/GUIDELINES.md](./metrics/GUIDELINES.md) -- OpenTelemetry metrics, Prometheus, Grafana alignment
- [docs/dev-guides/](./docs/dev-guides/) -- Development guides (e.g., HBI testing)
- [docs/runbooks/](./docs/runbooks/) -- Operational runbooks for troubleshooting and incident response

### To Build:
`make build`

### To Build Container Image:

Log in to the required registries, then run the `docker-build-push` target with your image destination. The base image is pulled from `registry.access.redhat.com` during the build, so both logins are required.

    podman login quay.io                     # or: docker login quay.io
    podman login registry.access.redhat.com  # or: docker login registry.access.redhat.com

    make docker-build-push IMAGE=quay.io/your-org/inventory-consumer

If the build fails with an authentication error mentioning `registry.access.redhat.com`, ensure you are logged in using your Red Hat Customer Portal credentials. See the [Red Hat Registry Authentication guide](https://access.redhat.com/RegistryAuthentication) for details.

`podman` is used automatically if available, otherwise `docker` is used. Override with `DOCKER=docker make docker-build-push IMAGE=quay.io/your-org/inventory-consumer`.

### To Run:

Prerequisites: You need to have the basic kafka setup deployed in order to test. You can use any of the docker compose options ([standard](https://github.com/project-kessel/inventory-api/tree/main?tab=readme-ov-file#running-locally-using-docker-compose) or [more elaborate](https://github.com/project-kessel/inventory-api/blob/main/docs/dev-guides/docker-compose-options.md)) in Inventory API to setup the backend services, including Kafka

#### Using local binary

```shell
make build
./bin/inventory-consumer start --consumer.bootstrap-servers localhost:9092
```

**Using Podman (requires you build the image first)**
```shell
podman run --network kessel -d quay.io/YOUR-IMAGE-HERE:TAG start --consumer.bootstrap-servers kafka:9093
```

#### Using Podman Compose (Recommended)

>[!NOTE]
>The podman compose setup is meant to replicate an ephemeral-like environment locally and requires Inventory API and Relations API to be running by default. You will need both repos cloned down in order to set everything up.

1. In the root of your cloned Inventory API repo: `make inventory-up-relations-ready`

2. In the root of your cloned Relations API reo: `make relations-api-up`

3. Spin up the Kessel Inventory Consumer components
```shell
# Deploy KIC and related dependencies
# This will include a test HBI database for now, Kafka Connect cluster and topic creation
make inventory-consumer-up
```

This will allow you to test Kessel Inventory Consumer by producing messages to any created topics that the consumer is configured to monitor (see [config file](./development/configs/full-setup.yaml)).

See the [Development Docs](./docs/dev-guides/) for info on specific service provider testing use cases.

#### Using Ephemeral

```shell
# Deploy Kessel Services (this will also deploy relations and inventory api)
bonfire deploy kessel -C kessel-inventory-consumer
```

#### Testing in Ephemeral

>[!NOTE]
>Since the Kessel Inventory Consumer is only used for HBI currently, for any testing, its recommended to use the process in the insights-deployer-script for standing everything up and testing. Any below testing is just basic validation of the service working and is not indicative of a final setup. See the [HBI Migration Runbook](https://github.com/project-kessel/insights-service-deployer/blob/main/docs/hbi-migration-runbook.md) for more details

To perform a basic test in Ephemeral you need to produce a message to the topic for the consumer to create the resource

```shell
# Spin up the Kessel Debug container
oc process --local \
    -f https://raw.githubusercontent.com/project-kessel/inventory-api/refs/heads/main/tools/kessel-debug-container/kessel-debug-deploy.yaml \
    -p ENV="env-$(oc project)" | oc apply -f -

# rsh to the debug container
oc rsh kessel-debug

# Setup Kafka env vars
source /usr/local/bin/env-setup.sh

# Create an HBI Host using Outbox
echo '{"schema":{"type":"string","optional":false},"payload":"dd1b73b9-3e33-4264-968c-e3ce55b9afec"}|{"schema":{"type":"struct","fields":[{"type":"string","optional":true,"field":"type"},{"type":"string","optional":true,"field":"reporter_type"},{"type":"string","optional":true,"field":"reporter_instance_id"},{"type":"struct","fields":[{"type":"struct","fields":[{"type":"string","optional":true,"field":"local_resource_id"},{"type":"string","optional":true,"field":"api_href"},{"type":"string","optional":true,"field":"console_href"},{"type":"string","optional":true,"field":"reporter_version"}],"optional":true,"name":"metadata"},{"type":"struct","fields":[{"type":"string","optional":true,"field":"workspace_id"}],"optional":true,"name":"common"},{"type":"struct","fields":[{"type":"string","optional":true,"field":"satellite_id"},{"type":"string","optional":true,"field":"subscription_manager_id"},{"type":"string","optional":true,"field":"insights_id"},{"type":"string","optional":true,"field":"ansible_host"}],"optional":true,"name":"reporter"}],"optional":true,"name":"representations"}],"optional":true,"name":"payload"},"payload":{"type":"host","reporter_type":"hbi","reporter_instance_id":"3088be62-1c60-4884-b133-9200542d0b3f","representations":{"metadata":{"local_resource_id":"dd1b73b9-3e33-4264-968c-e3ce55b9afec","api_href":"https://apiHref.com/","console_href":"https://www.console.com/","reporter_version":"2.7.16"},"common":{"workspace_id":"a64d17d0-aec3-410a-acd0-e0b85b22c076"},"reporter":{"satellite_id":"2c4196f1-0371-4f4c-8913-e113cfaa6e67","subscription_manager_id":"af94f92b-0b65-4cac-b449-6b77e665a08f","insights_id":"05707922-7b0a-4fe6-982d-6adbc7695b8f","ansible_host":"host-1"}}}}' | kcat -P -b $BOOTSTRAP_SERVERS -H "operation=ReportResource" -H "version=v1beta2" -t outbox.event.hbi.hosts -K "|"

# Delete the same HBI Host using Outbox
echo '{"schema":{"type":"string","optional":false},"payload":"dd1b73b9-3e33-4264-968c-e3ce55b9afec"}|{"schema":{"type":"struct","fields":[{"type":"struct","fields":[{"type":"string","optional":true,"field":"resource_type"},{"type":"string","optional":true,"field":"resource_id"},{"type":"struct","fields":[{"type":"string","optional":true,"field":"type"}],"optional":true,"name":"reporter"}],"optional":true,"name":"reference"}],"optional":true,"name":"payload"},"payload":{"reference":{"resource_type":"host","resource_id":"dd1b73b9-3e33-4264-968c-e3ce55b9afec","reporter":{"type":"hbi"}}}}' | kcat -P -b $BOOTSTRAP_SERVERS -H "operation=DeleteResource" -H "version=v1beta2" -t outbox.event.hbi.hosts -K "|"
```

### Monitoring

Prometheus metrics can be captured from both the Kessel Inventory Consumer, and if deployed, the Kessel Kafka Connect pod

KIC metrics are available on port 9000:

```shell
# Run local binary/container or `oc port-forward svc/kessel-inventory-consumer-service 9000:9000`
curl localhost:9000/metrics
```

Kafka Connect metrics are available on port 9404:
```shell
oc port-forward kessel-kafka-connect-connect-0 9404:9404
curl localhost:9404/metrics
```

### Monitoring in Ephemeral using Podman Compose

The `monitoring-only` option available in Kessel Inventory's compose setup can be used to monitor replication-related workloads deployed in Ephemeral for performance testing.

To get started you'll need to deploy Kessel to ephemeral: `bonfire deploy kessel -C kessel-inventory-consumer`

Once deployed, follow the steps in the [guide](https://github.com/project-kessel/inventory-api/blob/main/docs/dev-guides/docker-compose-options.md#monitoring-stack-only) to setup the monitoring stack.
