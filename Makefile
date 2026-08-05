.PHONY: up down logs migrate seed

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

migrate:
	@docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U rtdigital -d rtdigital \
		-c "CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())"
	@for file in infrastructure/migrations/*.up.sql; do \
		version=$$(basename "$$file" .up.sql); \
		applied=$$(docker compose exec -T postgres psql -Atq -U rtdigital -d rtdigital \
			-c "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = '$$version')"); \
		if [ "$$applied" = "t" ]; then \
			continue; \
		fi; \
		echo "Applying $$file"; \
		{ printf 'BEGIN;\n'; cat "$$file"; \
		  printf "INSERT INTO schema_migrations (version) VALUES ('%s');\nCOMMIT;\n" "$$version"; } | \
			docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U rtdigital -d rtdigital -f -; \
	done

seed: migrate
	@for file in infrastructure/seeds/*.sql; do \
		test -e "$$file" || continue; \
		echo "Seeding $$file"; \
		docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U rtdigital -d rtdigital -f - < "$$file"; \
	done
	@echo "Seeding development super admin"
	@docker compose exec -T api go run ./cmd/seed-super-admin
	@echo "Seeding development demo data"
	@docker compose exec -T api go run ./cmd/seed-demo
