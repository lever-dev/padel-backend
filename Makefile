.PHONY: docs
docs:
	swag init -g ./cmd/main.go -o=./docs --generatedTime=false --parseDependency --parseInternal

.PHONY: run_containers
run_containers:
	docker compose -f ./deployments/docker-compose.yml up -d

.PHONY: stop_containers
stop_containers:
	docker compose -f ./deployments/docker-compose.yml down

.PHONY: run_server
run_server:
	go run ./cmd/. serve

.PHONY: tests
tests:
	docker compose -f deployments/docker-compose.yml --profile tests up --build tests --exit-code-from tests

.PHONY: logs
logs:
	docker compose -f deployments/docker-compose.yml logs

.PHONY: fmt
fmt:
	@echo "🧹 Formatting Go code..."
	@gofmt -l -w `find . -type f -name '*.go' -not -path "./vendor/*"`
	@golines --max-len=120 --base-formatter=gofmt --shorten-comments --ignore-generated  --ignored-dirs=vendor -w .
	@echo "✅ Code formatted successfully"

.PHONY: lint
lint:
	golangci-lint run
