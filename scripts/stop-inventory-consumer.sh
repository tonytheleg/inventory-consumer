#!/bin/bash
set -e

export CONFIG=""

# Function to check if a command is available
source ./scripts/check_docker_podman.sh
${DOCKER} compose -f development/docker-compose.yaml down
