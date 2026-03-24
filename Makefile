.PHONY: build build-server build-cli test lint clean

build: build-cli build-server

build-cli:
	go build -o bin/kit ./cmd/kit

build-server:
	go build -o bin/kitd ./cmd/kitd

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin/

docker:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down
