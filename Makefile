test:
	@go clean -testcache
	@go test -cover -race ./...

.PHONY: db
db:
	@docker run -d \
		-p 5432:5432 \
		-e POSTGRES_PASSWORD=password \
		--name money-tracker-telegram \
		postgres:15-alpine
	@timeout 30 bash -c "until docker exec money-tracker-telegram pg_isready; do sleep 2; done"
