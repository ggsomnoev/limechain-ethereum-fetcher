DB_CONN_URL="pgx5://ethuser:ethpass@localhost:5432/postgres"

.PHONY: test
test:
	ginkgo run -race ./..

PHONY: run-docker
run-docker:
	docker compose down --volumes --remove-orphans 
	docker-compose up -d --build 

.PHONY: migrate-up
migrate-up:
	migrate -path ./migrations -database $(DB_CONN_URL) up

.PHONY: migrate-down
migrate-down:
	migrate -path ./migrations -database $(DB_CONN_URL) down