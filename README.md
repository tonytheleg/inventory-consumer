# PoC: Inventory Resource Consumer

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

Prerequisites: You need to have the basic kafka setup deployed in order to test. You can use the [split setup](https://github.com/project-kessel/inventory-api/blob/main/docs/dev-guides/docker-compose-options.md#local-kessel-inventory--docker-compose-infra-split-setup) target in Inventory API to setup the backend services

**Using local binary**

```shell
make build
./bin/inventory-consumer start --consumer.bootstrap-servers localhost:9092
```

**Using Podman (requires you build the image first)**
```shell
podman run --network kessel -d quay.io/YOUR-IMAGE-HERE:TAG start --consumer.bootstrap-servers kafka:9093
```

