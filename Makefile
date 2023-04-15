GRPC_GEN_FILES=./proto/notifications.proto ./proto/chat_updates.proto

# Used in Dockerfile.dev for live reloading
start:
	./build/app --host 0.0.0.0 --port 80

# Generates all grpc stuff
generate:
	protoc --go_out=. --go_opt=paths=import --go-grpc_out=. --go-grpc_opt=paths=import $(GRPC_GEN_FILES)

build: generate
	go build -o ./bin/app cmd/main.go

run: build
	docker-compose up -d

coverage:
	go test -short -count=1 -race -coverprofile=coverage.out ./...
	go tool cover -html="coverage.out"
	rm coverage.out

all: generate run
