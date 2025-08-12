# Local HBI Outbox Consumer Processing Using Kafka Event

This process is useful for testing the Kessel service individually and does not require any extra HBI bits. This process works by publishing a message to the HBI outbox topic which will then be captured by the Inventory Consumer and replicated down to relations.

To publish the messages, you will need the `kcat` cli (See [Install instructions](https://github.com/edenhill/kcat?tab=readme-ov-file#install))

### Steps:

1. Spin up everything via podman compose (See [Using Podman Compose](../../README.md#using-podman-compose-recommended))

2. Publish messages using `kcat`

```shell
# Create an HBI Host using Outbox
echo '{"schema":{"type":"string","optional":false},"payload":"dd1b73b9-3e33-4264-968c-e3ce55b9afec"}|{"schema":{"type":"struct","fields":[{"type":"string","optional":true,"field":"type"},{"type":"string","optional":true,"field":"reporter_type"},{"type":"string","optional":true,"field":"reporter_instance_id"},{"type":"struct","fields":[{"type":"struct","fields":[{"type":"string","optional":true,"field":"local_resource_id"},{"type":"string","optional":true,"field":"api_href"},{"type":"string","optional":true,"field":"console_href"},{"type":"string","optional":true,"field":"reporter_version"}],"optional":true,"name":"metadata"},{"type":"struct","fields":[{"type":"string","optional":true,"field":"workspace_id"}],"optional":true,"name":"common"},{"type":"struct","fields":[{"type":"string","optional":true,"field":"satellite_id"},{"type":"string","optional":true,"field":"subscription_manager_id"},{"type":"string","optional":true,"field":"insights_inventory_id"},{"type":"string","optional":true,"field":"ansible_host"}],"optional":true,"name":"reporter"}],"optional":true,"name":"representations"}],"optional":true,"name":"payload"},"payload":{"type":"host","reporter_type":"hbi","reporter_instance_id":"3088be62-1c60-4884-b133-9200542d0b3f","representations":{"metadata":{"local_resource_id":"dd1b73b9-3e33-4264-968c-e3ce55b9afec","api_href":"https://apiHref.com/","console_href":"https://www.console.com/","reporter_version":"2.7.16"},"common":{"workspace_id":"a64d17d0-aec3-410a-acd0-e0b85b22c076"},"reporter":{"satellite_id":"2c4196f1-0371-4f4c-8913-e113cfaa6e67","subscription_manager_id":"af94f92b-0b65-4cac-b449-6b77e665a08f","insights_inventory_id":"05707922-7b0a-4fe6-982d-6adbc7695b8f","ansible_host":"host-1"}}}}' | kcat -P -b localhost:9092 -H "operation=ReportResource" -H "version=v1beta2" -t outbox.event.hbi.hosts -K "|"

# Delete the same HBI Host using Outbox
echo '{"schema":{"type":"string","optional":false},"payload":"dd1b73b9-3e33-4264-968c-e3ce55b9afec"}|{"schema":{"type":"struct","fields":[{"type":"struct","fields":[{"type":"string","optional":true,"field":"resource_type"},{"type":"string","optional":true,"field":"resource_id"},{"type":"struct","fields":[{"type":"string","optional":true,"field":"type"}],"optional":true,"name":"reporter"}],"optional":true,"name":"reference"}],"optional":true,"name":"payload"},"payload":{"reference":{"resource_type":"host","resource_id":"dd1b73b9-3e33-4264-968c-e3ce55b9afec","reporter":{"type":"hbi"}}}}' | kcat -P -b localhost:9092 -H "operation=DeleteResource" -H "version=v1beta2" -t outbox.event.hbi.hosts -K "|"
```

Once sent, you can review the logs or any databases and see the replication throughout

```shell
# check Inventory Consumer logs
podman logs development-inventory-consumer-1

# check Inventory API logs for resource creation and internal consumer replication
podman logs development-inventory-api-1

# check Relations API logs for tuple creation events
podman logs relations-api-relations-api-1

# access resources in Inventory API DB
psql -h localhost -p 5433 -d spicedb -U postgres # requires password available in Inventory API repo
```



# Local HBI Migration Testing using Hosts Table and Debezium

### Steps:

1. Spin up everything via podman compose (See [Using Podman Compose](../../README.md#using-podman-compose-recommended))
2. Configure the HBI database: `make setup-hbi-db`
3. Generate SQL import files with host records using the [db-generator](https://github.com/tonytheleg/db-generator)
4. Import host records: `PGPASSWORD=supersecurewow psql -h localhost -p 5435 -d host-inventory -U postgres -f path/to/import-sql-files`
5. Setup the Connectors: `make setup-connectors`

At this point both the migration and outbox connecters are started and Debezium will begin the snapshot of the hosts table and capture all existing records.

To test the outbox, you can import outbox records generated using the same db-generator tool and process:

`PGPASSWORD=supersecurewow psql -h localhost -p 5435 -d host-inventory -U postgres -f path/to/outbox-import-sql-files`


### Clean Up:

1. Shut down the KIC setup: `make inventory-consumer-down`
2. Shut down Kessel Inventory: `cd path/to/inventory-api && make inventory-down`
3. Shut down Kessel Relations: `cd path/to/relations-api && make relations-api-down`


### Incremental Snapshots

To test using incremental snapshots:

1. Spin up everything via podman compose (See [Using Podman Compose](../../README.md#using-podman-compose)
2. Configure the HBI database: `make setup-hbi-db`
3. Generate SQL import files with host records using the [db-generator](https://github.com/tonytheleg/db-generator)
4. Import host records: `PGPASSWORD=supersecurewow psql -h localhost -p 5435 -d host-inventory -U postgres -f path/to/import-sql-files`
5. Setup the `no-snapshot` Connector: `make setup-migration-connector-no-snapshot`

When the connector is started, the intial snapshot will not run due to `snapshot.mode` being set to `"no_data"`.To trigger a snapshot, you must produce a signal to the signal topic which will trigger the snapshot. Note, the signal table is required even when using the topic as it leverages the table as part of its snapshot process

To trigger the snapshot (requires [kcat](https://github.com/edenhill/kcat?tab=readme-ov-file#install)):

`echo 'host-inventory|{"type":"execute-snapshot","data":{"data-collections":["hbi.hosts"],"type":"INCREMENTAL"}}' | kcat -P -b localhost:9092 -t host-inventory.signal -K "|"`
