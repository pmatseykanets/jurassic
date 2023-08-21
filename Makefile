BINARY=jurassic

default: test-unit

dep:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

clean:
	rm -f ./${BINARY}

test-unit:
	go test -race -coverprofile=coverage.txt -covermode=atomic --tags=unit ./...

test-integration:
	go test -race -coverprofile=coverage.txt -covermode=atomic --tags=integration ./...

build: clean
	CGO_ENABLED=0 go build -o ${BINARY} -ldflags "-s -w -X 'main.buildVersion=# Built $(shell date -u -R) with $(shell go version) at $(shell git rev-parse HEAD)' -X 'main.version=$(shell git describe --tags --always --dirty --match "v[0-9]*" --abbrev=4 | sed -e 's/^v//')'"

run:
	go run main.go

.PHONY: dep, clean, test-unit, test-integration, build, run
