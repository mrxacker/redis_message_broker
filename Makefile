APP_NAME := redis_message_broker
BIN_DIR := bin
COMPOSE_FILE=docker-compose.override.yml
ENV_FILE=.env

ifneq (,$(wildcard .env))
	include .env
	export
endif

.PHONY: run build test clean

run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

up:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) up -d --build

down:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) down -v
