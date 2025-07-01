PQ_CONTAINER = postgres

gen-proto: ## Generate protobuf files
	@echo "Generating protobuf files..."
	@mkdir -p proto-gen
	protoc --go_out=proto-gen --go_opt=paths=source_relative \
		--go-grpc_out=proto-gen --go-grpc_opt=paths=source_relative \
		proto/translation/translation.proto proto/validation/validation.proto

db-start: ## Start postgres server on a docker container
	@docker run --rm --name $(PQ_CONTAINER) \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=postgres \
		-p 5432:5432 \
		-d \
		postgres


db-stop: ## Stop the database container if running
	@if [ ! "$(shell docker ps | grep '$(PQ_CONTAINER)' )" = "" ]; then \
		docker stop postgres; \
	fi

run-server: ## Run the server
	@echo "Running server..."
	go run main.go

run-all: db-stop sleep db-start sleep run-server ## Run the server with a database

sleep:
	@sleep 5