# Local HBI Migration Testing using Hosts Table


### Steps:

1. Spin up everything via podman compose (See [Using Podman Compose](../../README.md#using-podman-compose)
2. Configure the HBI database: `make setup-hbi-db`
3. Generate SQL import files with host records using the [db-host-generator.sh](../../scripts/db-host-generator.sh) script
4. Import host records: `PGPASSWORD=supersecurewow psql -h localhost -p 5432 -d host-inventory -U postgres -f path/to/import-sql-files
5. Setup the Connector: `make setup-migration-connector`

At this point the connecter is started, and Debezium will begin the snapshot of the hosts table and capture all existing records.


### Clean Up:

1. Shut down the KIC setup: `make inventory-consumer-down`
2. Shut down Kessel Inventory: `cd path/to/inventory-api && make inventory-down`
3. Shut down Kessel Relations: `cd path/to/relations-api && make relations-api-down`
