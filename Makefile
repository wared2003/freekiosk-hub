# Variables
BINARY_NAME=freekiosk-hub
MAIN_PATH=cmd/server/main.go
DB_NAME=freekiosk.db

.PHONY: all build run clean help deps

## all: Compile le projet
all: build

## build: Compile le binaire dans le dossier bin/
build:
	@echo "🔨 Compilation de $(BINARY_NAME)..."
	@mkdir -p bin
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

## run: Lance l'application directement avec go run
run:
	@echo "🚀 Démarrage de FreeKiosk Hub..."
	go run $(MAIN_PATH)

## deps: Nettoie et télécharge les dépendances Go
deps:
	@echo "📦 Mise à jour des dépendances..."
	go mod tidy
	go mod download

## clean: Supprime le binaire et la base de données locale
clean:
	@echo "🧹 Nettoyage..."
	@rm -rf bin/
	@if [ -f $(DB_NAME) ]; then rm $(DB_NAME); echo "🗑️ Base de données supprimée"; fi

## help: Affiche cette aide
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'