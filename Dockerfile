FROM registry.access.redhat.com/ubi9/ubi-minimal:9.6-1752587672 AS builder

ARG TARGETARCH
USER root
RUN microdnf install -y tar gzip make which gcc gcc-c++ cyrus-sasl-lib findutils git go-toolset

WORKDIR /workspace

COPY go.mod go.sum ./

ENV CGO_ENABLED 1
RUN go mod download

COPY cmd ./cmd
COPY consumer ./consumer
COPY internal ./internal
COPY metrics ./metrics
COPY main.go Makefile ./

ARG VERSION
RUN VERSION=${VERSION} make build

FROM registry.access.redhat.com/ubi9/ubi-minimal:9.6-1752587672

COPY --from=builder /workspace/bin/inventory-consumer /usr/local/bin/

USER 1001
ENV PATH="$PATH:/usr/local/bin"
ENTRYPOINT ["inventory-consumer"]
CMD ["start"]
