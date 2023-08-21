BINARY=jurassic

default: test

dep:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

clean:
	rm -f ./${BINARY}

test:
	go vet ./...
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

build: clean
	CGO_ENABLED=0 go build -o ${BINARY} -ldflags "-s -w -X 'main.buildVersion=# Built $(shell date -u -R) with $(shell go version) at $(shell git rev-parse HEAD)' -X 'main.version=$(shell git describe --tags --always --dirty --match "v[0-9]*" --abbrev=4 | sed -e 's/^v//')'"

run:
	go run main.go

.PHONY: dep, clean, test, build, run
