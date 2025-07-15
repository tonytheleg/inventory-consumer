#!/bin/bash

# prints a handy help menu with usage
help_me() {
    echo "USAGE: db-host-generator.sh {-n <NUM_HOSTS>} [-c <NUM_FILES>|-h]"
    echo "db-host-generator: creates unique fake host records to add to HBI database for migration testing"
    echo "The records are created in batches of 1000 records and dumped to a file for importing"
    echo ""
    echo "REQUIRED ARGUMENTS:"
    echo "  -n NUM_HOSTS: The number of hosts to create in 1000's (ie -n 1 creates 1000 hosts)"
    echo ""
    echo "OPTIONS:"
    echo "  -h Prints usage information"
    echo "  -c NUM_HOSTS: If provided, creates X number of files with -n number hosts*1000 per file"

    echo ""
    echo "EXAMPLE:"
    echo "Create 5000 hosts in a single import file"
    echo "  load-generator.sh -n 5"
    echo ""
    echo "Create 5 files with 5000 hosts in each file (total of 25,000 hosts)"
    echo "  load-generator.sh -n 5 -c 5"
    exit 0
}

generate_import_file() {
    FILE_NAME=$1
    for ((i = 1  ; i <= ${NUM_BATCHES} ; i++)); do
        # Generate a unique ID and hostname
        ID=$(printf "'%s'" $(uuidgen))
        HOST=$(printf "'%s'" $(echo ${ID:1:7}.foo.redhat.com))
        CANONICAL_FACTS="'{\"insights_id\": \"62a6917d-bff6-4b77-a185-d570472b0699\", \"provider_id\": \"26be74db-d31e-4d94-bde1-4a0914901d98\", \"provider_type\": \"ibm\"}'"

        if [[ $(( $i%1000 )) -eq 0 ]]; then
            NEW_VALUE="($ID, $HOST, '2025-07-10 15:03:41.855142+00', '2025-07-10 15:03:41.855145+00', $CANONICAL_FACTS::jsonb, '2025-07-11 20:03:41.75047+00', 'rhsm-conduit', '321', '[]', '2025-07-10 15:03:41.75047+00', '2025-07-17 15:03:41.75047+00', '2025-07-24 15:03:41.75047+00');"
            VALUES+=($NEW_VALUE)
            echo "INSERT INTO hbi.hosts (id, display_name, created_on, modified_on, canonical_facts, stale_timestamp, reporter, org_id, groups, last_check_in, stale_warning_timestamp, deletion_timestamp) VALUES ${VALUES[*]}" >> $FILE_NAME
            echo "" >> $FILE_NAME
            unset VALUES
        else
            NEW_VALUE="($ID, $HOST, '2025-07-10 15:03:41.855142+00', '2025-07-10 15:03:41.855145+00', $CANONICAL_FACTS::jsonb, '2025-07-11 20:03:41.75047+00', 'rhsm-conduit', '321', '[]', '2025-07-10 15:03:41.75047+00', '2025-07-17 15:03:41.75047+00', '2025-07-24 15:03:41.75047+00'),"
            VALUES+=($NEW_VALUE)
        fi
    done
}

while getopts "n:c:h" flag; do
    case "${flag}" in
        n) NUM_HOSTS=${OPTARG};;
        c) NUM_FILES=${OPTARG};;
        h) help_me;;
    esac
done

if [[  -z "${NUM_HOSTS}" ]]; then
  echo "Error: required arguments not provided"
  help_me
fi

declare -a VALUES
NUM_BATCHES=$(( $NUM_HOSTS*1000 ))

rm import-*.sql

for ((count = 1  ; count <= ${NUM_FILES} ; count++)); do
    generate_import_file import-${count}.sql
done
