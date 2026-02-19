run:
    APP_PROFILE=dev go run ./cmd/app

test:
    go test ./...

build:
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/app ./cmd/app

docker:
    docker build -t goboot-app:latest .