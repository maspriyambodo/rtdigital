.PHONY: up down logs migrate seed

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

migrate:
	@for file in infrastructure/migrations/*.up.sql; do \
		echo "Applying $$file"; \
		docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U rtdigital -d rtdigital -f - < "$$file"; \
	done

seed: migrate
	@for file in infrastructure/seeds/*.sql; do \
		test -e "$$file" || continue; \
		echo "Seeding $$file"; \
		docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U rtdigital -d rtdigital -f - < "$$file"; \
	done