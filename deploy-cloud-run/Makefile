dev:
	docker compose -f docker-compose-dev.yml up -d

prod:
	docker compose -f docker-compose-prod.yml up -d --build

stop-dev:
	docker compose -f docker-compose-dev.yml down

stop-prod:
	docker compose -f docker-compose-prod.yml down

test:
	go test -v ./...

.PHONY: dev prod stop-dev stop-prod test