.PHONY: test
test:
	@echo ""
	go test ./... -count=1 -race -short -covermode=atomic -coverprofile=coverage.txt
	@echo "Overall test coverage:"
	go tool cover -func=coverage.txt | grep total: | awk '{print $$3}'