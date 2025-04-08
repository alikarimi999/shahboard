COMPOSE_CMD = docker-compose -f docker-compose.base.yml

SERVICES = $(shell ls deploy)
PROFILE_PATH_ALL = $(patsubst %,-f deploy/%/production/docker-compose.yml,$(SERVICES))

up:
	@if [ -z "$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		$(COMPOSE_CMD) $(PROFILE_PATH_ALL) up -d; \
	else \
		$(COMPOSE_CMD) -f deploy/$(filter-out $@,$(MAKECMDGOALS))/production/docker-compose.yml up -d; \
	fi


down:
	@if [ -z "$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		$(COMPOSE_CMD) $(PROFILE_PATH_ALL) down; \
	else \
		SERVICE=$(filter-out $@,$(MAKECMDGOALS)); \
		if [ "$$SERVICE" = "wsgateway" ]; then \
			$(COMPOSE_CMD) -f deploy/$$SERVICE/production/docker-compose.yml down $$SERVICE; \
		else \
			$(COMPOSE_CMD) -f deploy/$$SERVICE/production/docker-compose.yml down $$SERVICE-service; \
		fi; \
	fi

stop:
	@if [ -z "$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		$(COMPOSE_CMD) $(PROFILE_PATH_ALL) stop; \
	else \
		SERVICE=$(filter-out $@,$(MAKECMDGOALS)); \
		if [ "$$SERVICE" = "wsgateway" ]; then \
			$(COMPOSE_CMD) -f deploy/$$SERVICE/production/docker-compose.yml stop $$SERVICE; \
		else \
			$(COMPOSE_CMD) -f deploy/$$SERVICE/production/docker-compose.yml stop $$SERVICE-service; \
		fi; \
	fi


build:
	@if [ -z "$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		$(COMPOSE_CMD) $(PROFILE_PATH_ALL) build; \
	else \
		$(COMPOSE_CMD) -f deploy/$(filter-out $@,$(MAKECMDGOALS))/production/docker-compose.yml build; \
	fi

logs:
	@if [ -z "$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		echo "Please specify a service name (e.g., 'make logs auth')"; \
	else \
		SERVICE=$(filter-out $@,$(MAKECMDGOALS)); \
		if [ "$$SERVICE" = "wsgateway" ]; then \
			$(COMPOSE_CMD) -f deploy/$$SERVICE/production/docker-compose.yml logs -f $$SERVICE; \
		else \
			$(COMPOSE_CMD) -f deploy/$$SERVICE/production/docker-compose.yml logs -f $$SERVICE-service; \
		fi; \
	fi

start-%:
	$(COMPOSE_CMD) -f deploy/$*/production/docker-compose.yml up -d

stop-%:
	$(COMPOSE_CMD) -f deploy/$*/production/docker-compose.yml stop

logs-%:
	$(COMPOSE_CMD) -f deploy/$*/production/docker-compose.yml logs -f

restart-%:
	make stop-$*
	make start-$*

ps:
	docker ps

%:
	@: