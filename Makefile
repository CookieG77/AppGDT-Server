# Commandes courantes du projet. Utilisation : make <cible> (make help pour la liste).
# Fonctionne sous Linux, macOS et Windows (make installé via Chocolatey, Scoop ou winget).

ifeq ($(OS),Windows_NT)
	EXT := .exe
	RM_BIN := if exist bin rmdir /s /q bin
else
	EXT :=
	RM_BIN := rm -rf bin
endif

BIN := bin/gdt-server$(EXT)

.PHONY: help db-up db-down run seed seed-reset build test vet fmt clean

help:
	@echo Cibles disponibles :
	@echo   make db-up       - demarre PostgreSQL (docker compose)
	@echo   make db-down     - arrete PostgreSQL
	@echo   make run         - lance l API
	@echo   make seed        - cree les comptes de demonstration
	@echo   make seed-reset  - recree les comptes de demonstration
	@echo   make build       - compile l API dans bin/
	@echo   make test        - lance les tests
	@echo   make vet         - analyse statique du code
	@echo   make fmt         - formate le code
	@echo   make clean       - supprime bin/

db-up:
	docker compose up -d

db-down:
	docker compose down

run:
	go run ./cmd/server

seed:
	go run ./cmd/seed

seed-reset:
	go run ./cmd/seed -reset

build:
	go build -o $(BIN) ./cmd/server

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

clean:
	$(RM_BIN)
