.PHONY: build up down test quicktest

build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

test: 
	go test -v
