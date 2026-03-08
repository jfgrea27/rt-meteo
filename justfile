set dotenv-load

# List available recipes
default:
    @just --list

# Run all Go tests
test:
    cd src && go test ./...

# Build a specific service (api, consumer, producer)
build SVC:
    cd src && go build -o ../bin/{{SVC}} ./cmd/{{SVC}}

docker-build SVC:
    cd src && docker build --build-arg SVC={{SVC}} -t {{SVC}}:latest .

docker-run SVC: 
    just docker-build {{SVC}}
    docker run -it --rm {{SVC}}:latest

docker-up:
    cd src && docker compose down -v
    cd src && docker compose up --build

# Build all services
build-all:
    just build api
    just build consumer
    just build producer
