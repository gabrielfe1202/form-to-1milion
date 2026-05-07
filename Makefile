# Makefile para o projeto form-to-1milion
# Use:
#   make up         - sobe os containers em segundo plano
#   make down       - para e remove os containers
#   make restart    - reinicia os containers
#   make status     - mostra status dos containers
#   make logs       - mostra logs do docker-compose
#   make test       - executa scenario de carga (se o k6 estiver configurado)

COMPOSE=docker-compose

up:
	$(COMPOSE) up -d

up-and-build:
	$(COMPOSE) up -d --build

build:
	$(COMPOSE) build

create-queue:
	aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name form-to-1milion-queue

build-start: up-and-build create-queue

start: up create-queue

down:
	$(COMPOSE) down

restart: down up

status:
	$(COMPOSE) ps

logs:
	$(COMPOSE) logs -f

# Executa teste k6 localmente; exige k6 instalado na máquina.
test:
	k6 run tests/load/load-test.js

# Alias comum
run: up
	echo "Containers iniciados"
