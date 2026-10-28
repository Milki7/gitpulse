.PHONY: build run test vet fmt-check

build:
	go build ./...

run:
	go run ./cmd/gitpulse

test:
	go test -race ./...

vet:
	go vet ./...

fmt-check:
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

