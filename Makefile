-include .env

# Apply the contents of the .env to the terminal, so that the docker-compose file can use them in its builds
export $(shell sed 's/=.*//' .env)

gomi:
	@echo "==============================================="
	@echo "Make: gomi - running gomi"
	@echo "==============================================="
	@docker compose up -d db gomi

close:
	@echo "==============================================="
	@echo "Make: close - closing Docker containers"
	@echo "==============================================="
	@docker compose down

clean: ## Closes and cleans (removes) all project containers
	@echo "==============================================="
	@echo "Make: clean - closing and cleaning Docker containers"
	@echo "==============================================="
	@docker compose down -v --rmi all --remove-orphans

