# Kessel Inventory Consumer

The Kessel Inventory Consumer (KIC) is a standalone dedicated Kafka consumer group used to expose an eventing based entry point to the Kessel Inventory API. Its purpose is to subscribe to Service Provider owned Kafka topics and ensure reporter resource updates are replicated to Inventory API through events.

### To Build:
`make local-build`

### To Build Container Image:

_Linux/Windows_
```shell
export IMAGE=your-quay-repo
make docker-build-push
```

_MacOS_

```shell
export QUAY_REPO_INVENTORY=your-quay-repo # required
podman login quay.io # required, this target assumes you are already logged in
make build-push-minimal
```

### To Run:

Prerequisites: You need to have the basic kafka setup deployed in order to test. You can use any of the docker compose options ([standard](https://github.com/project-kessel/inventory-api/tree/main?tab=readme-ov-file#running-locally-using-docker-compose) or [more elaborate](https://github.com/project-kessel/inventory-api/blob/main/docs/dev-guides/docker-compose-options.md)) in Inventory API to setup the backend services, including Kafka

**Using local binary**

```shell
make build
./bin/inventory-consumer start --consumer.bootstrap-servers localhost:9092
```

**Using Podman (requires you build the image first)**
```shell
podman run --network kessel -d quay.io/YOUR-IMAGE-HERE:TAG start --consumer.bootstrap-servers kafka:9093
```

**Using Ephemeral**

```shell
# Deploy Kessel Services (this will also deploy relations and inventory api)
bonfire deploy kessel -C kessel-inventory-consumer

# Once its all running, update the spicedb schema for one that supports HBI
oc apply -f https://gist.githubusercontent.com/akoserwal/a061a2959862caa653aa8c8836db874b/raw/7cd73fd045ed8f30c850349bb7ff3264b2d35c8e/spicedb-schema-configmap.yaml

# Kick the relations pod to load the new schema
oc delete pod -l app=kessel-relations

# To test in Ephemeral you need to produce a message to the topic for the consumer to create the resource
# You can test with my personal kcat image
BOOTSTRAP_SERVERS=$(oc get secret kessel-inventory -o json | jq -r '.data."cdappconfig.json"' | base64 -d | jq -r '.kafka.brokers[] | "\(.hostname):\(.port)"')
oc run kcat --rm -i --tty --image quay.io/anatale/kcat:fedora --env BOOTSTRAP_SERVERS="$BOOTSTRAP_SERVERS" -- bash

# Create an HBI Host
echo '{"schema":{"type":"string","optional":false},"payload":"dd1b73b9-3e33-4264-968c-e3ce55b9afec"}|{"schema":{"type":"struct","fields":[{"type":"string","optional":true,"field":"type"},{"type":"string","optional":true,"field":"reporter_type"},{"type":"string","optional":true,"field":"reporter_instance_id"},{"type":"struct","fields":[{"type":"struct","fields":[{"type":"string","optional":true,"field":"local_resource_id"},{"type":"string","optional":true,"field":"api_href"},{"type":"string","optional":true,"field":"console_href"},{"type":"string","optional":true,"field":"reporter_version"}],"optional":true,"name":"metadata"},{"type":"struct","fields":[{"type":"string","optional":true,"field":"workspace_id"}],"optional":true,"name":"common"},{"type":"struct","fields":[{"type":"string","optional":true,"field":"satellite_id"},{"type":"string","optional":true,"field":"subscription_manager_id"},{"type":"string","optional":true,"field":"insights_inventory_id"},{"type":"string","optional":true,"field":"ansible_host"}],"optional":true,"name":"reporter"}],"optional":true,"name":"representations"}],"optional":true,"name":"payload"},"payload":{"type":"host","reporter_type":"hbi","reporter_instance_id":"3088be62-1c60-4884-b133-9200542d0b3f","representations":{"metadata":{"local_resource_id":"dd1b73b9-3e33-4264-968c-e3ce55b9afec","api_href":"https://apiHref.com/","console_href":"https://www.console.com/","reporter_version":"2.7.16"},"common":{"workspace_id":"a64d17d0-aec3-410a-acd0-e0b85b22c076"},"reporter":{"satellite_id":"2c4196f1-0371-4f4c-8913-e113cfaa6e67","subscription_manager_id":"af94f92b-0b65-4cac-b449-6b77e665a08f","insights_inventory_id":"05707922-7b0a-4fe6-982d-6adbc7695b8f","ansible_host":"host-1"}}}}' | kcat -P -b $BOOTSTRAP_SERVERS -H "operation=created" -t hbi.replication.events -K "|"

# Delete the same HBI Host
echo '{"schema":{"type":"string","optional":false},"payload":"dd1b73b9-3e33-4264-968c-e3ce55b9afec"}|{"schema":{"type":"struct","fields":[{"type":"struct","fields":[{"type":"string","optional":true,"field":"resource_type"},{"type":"string","optional":true,"field":"resource_id"},{"type":"struct","fields":[{"type":"string","optional":true,"field":"type"}],"optional":true,"name":"reporter"}],"optional":true,"name":"reference"}],"optional":true,"name":"payload"},"payload":{"reference":{"resource_type":"host","resource_id":"dd1b73b9-3e33-4264-968c-e3ce55b9afec","reporter":{"type":"hbi"}}}}' | kcat -P -b $BOOTSTRAP_SERVERS -H "operation=deleted" -t hbi.replication.events -K "|"
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

The monitoring stack available in [Kessel Inventory](https://github.com/project-kessel/inventory-api/blob/main/docs/dev-guides/docker-compose-options.md#monitoring-stack-only) can be used to monitor replication-related workloads in Ephemeral for performance testing.

The process consists of:
1. Starting the monitoring stack using podman (see above link)
2. Port forward each of the services locally

```shell
# Note: the address 0.0.0.0 is used to ensure the podman containers can use the special host.containers.internal address to access the host
oc port-forward --address 0.0.0.0 svc/kessel-inventory-api 8000:8000 &
oc port-forward --address 0.0.0.0 svc/kessel-inventory-consumer-service 9000:9000 &
oc port-forward --address 0.0.0.0 kessel-kafka-connect-connect-0 9404:9404 &
```

3. Access Prometheus in your browser at `localhost:9050`

Metrics will then be scraped from the running pods in ephemeral through your port-forwarding connection!
