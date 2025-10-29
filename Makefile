.PHONY: run_containers
run_containers:
	docker compose -f ./deployments/docker-compose.yml up -d

.PHONY: stop_containers
stop_containers:
	docker compose -f ./deployments/docker-compose.yml down

.PHONY: fmt
fmt:
	@echo "🧹 Formatting Go code..."
	@gofmt -s -w .
	@golines --max-len=120 --base-formatter=gofmt --shorten-comments --ignore-generated  --ignored-dirs=vendor -w .
	@echo "✅ Code formatted successfully"

