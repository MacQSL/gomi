-include .env

# Apply the contents of the .env to the terminal, so that the docker-compose file can use them in its builds
export $(shell sed 's/=.*//' .env)

postgres: ## Run postgres container
	@echo "==============================================="
	@echo "Make: postgres - run postgres container"
	@echo "==============================================="
	@docker compose up -d postgres

close: ## Closes all project containers
	@echo "==============================================="
	@echo "Make: close - closing Docker containers"
	@echo "==============================================="
	@docker compose down

clean: ## Closes and cleans (removes) all project containers
	@echo "==============================================="
	@echo "Make: clean - closing and cleaning Docker containers"
	@echo "==============================================="
	@docker compose down -v --rmi all --remove-orphans

