# Build stage -- the Go toolchain embeds the validated FIPS module in all binaries automatically.
FROM registry.access.redhat.com/hi/go:1.26-fips AS builder

WORKDIR /workspace

COPY go.mod go.sum ./

RUN go mod download

COPY cmd ./cmd
COPY consumer ./consumer
COPY internal ./internal
COPY metrics ./metrics
COPY main.go Makefile ./

ARG VERSION
RUN VERSION=${VERSION} make build

# Runtime stage -- set GODEBUG so the binary runs in FIPS mode.
FROM registry.access.redhat.com/hi/core-runtime:2.43-openssl-fips

WORKDIR /

COPY --from=builder /workspace/bin/inventory-consumer /usr/local/bin/

ENV GODEBUG=fips140=on

USER 1001
ENV PATH="$PATH:/usr/local/bin"
ENTRYPOINT ["inventory-consumer"]
CMD ["start"]
