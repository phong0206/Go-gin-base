SHELL := /bin/sh

APP_NAME := ticket-service
BIN_DIR := bin
BIN_PATH := $(BIN_DIR)/$(APP_NAME)
DOCKER_IMAGE ?= $(APP_NAME):local
AIR_BIN := $(shell go env GOPATH)/bin/air

.PHONY: setup deps tools env run dev build test test-race fmt vet tidy check clean docker-build docker-run seed

setup: deps tools env

deps:
	go mod download

tools:
	go install github.com/air-verse/air@v1.67.4

env:
	@test -f .env || cp .env.example .env

run:
	go run .

dev:
	$(AIR_BIN)

seed:
	go run database/seeds/admin.seed.go

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_PATH) .

test:
	go test ./...

test-race:
	go test -race ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

vet:
	go vet ./...

tidy:
	go mod tidy

check: fmt vet test

clean:
	rm -rf $(BIN_DIR)

docker-build:
	docker build -f docker/api.Dockerfile -t $(DOCKER_IMAGE) .

docker-run:
	docker run --rm --env-file .env -p 8080:8080 $(DOCKER_IMAGE)
