.PHONY: test lint run docker-build docker-local docker-prod swagger all

APP_NAME := personaforge-backend
IMAGE := personaforge-backend:latest

test:
	go test ./...

# Git Bash/MSYS on Windows rewrites `/app` -> `C:/Program Files/Git/app` unless we disable path conversion.
UNAME_S := $(shell uname -s 2>/dev/null || echo "")
ifneq (,$(findstring MINGW,$(UNAME_S)))
  DOCKER_ENV := MSYS_NO_PATHCONV=1 MSYS2_ARG_CONV_EXCL="*"
endif
ifneq (,$(findstring MSYS,$(UNAME_S)))
  DOCKER_ENV := MSYS_NO_PATHCONV=1 MSYS2_ARG_CONV_EXCL="*"
endif

lint:
	# Run golangci-lint for the whole module but with an increased timeout
	$(DOCKER_ENV) docker run --rm -v "$$(pwd):/app" -w /app golangci/golangci-lint:v1.56.2 golangci-lint run ./... --timeout=5m

swagger:
	# Regenerate Swagger JSON/YAML and docs package
	go run github.com/swaggo/swag/cmd/swag@v1.16.3 init -g main.go -o docs

run:
	go run .

docker-build:
	docker build -t $(IMAGE) .

docker-local:
	docker compose --env-file .env.local up --build

docker-prod:
	docker compose -f docker-compose.prod.yml --env-file .env.prod up --build -d

db-setup:
	@echo "Setting up database..."
	@bash setup-database.sh || powershell -ExecutionPolicy Bypass -File setup-database.ps1

all: test lint swagger docker-build docker-local


