.PHONY: build
build:
	mkdir -p bin/ && go build -ldflags "-X cmd.Version=$(VERSION)" -o ./bin/ ./...

.PHONY: test
test:
	@echo ""
	go test ./... -count=1 -race -short -covermode=atomic -coverprofile=coverage.txt
	@echo "Overall test coverage:"
	go tool cover -func=coverage.txt | grep total: | awk '{print $$3}'