dev:
	docker compose -f docker-compose-dev.yml up -d --build

prod:
	docker compose -f docker-compose-prod.yml up -d --build

stop:
	docker compose -f docker-compose-dev.yml down
	docker compose -f docker-compose-prod.yml down

.PHONY: dev prod stop