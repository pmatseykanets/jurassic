FROM golang:1.21 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o jurassic -ldflags "-s -w -X 'main.buildVersion=# Built $(date -u -R) with $(go version) at $(git rev-parse HEAD)' -X 'main.version=$(git describe --tags --always --dirty --match "v[0-9]*" --abbrev=4 | sed -e 's/^v//')'"

FROM gcr.io/distroless/static

USER nonroot:nonroot

COPY --from=builder --chown=nonroot:nonroot /src/jurassic /jurassic/jurassic
COPY --from=builder --chown=nonroot:nonroot /src/db/migrations /jurassic/db/migrations

CMD ["/jurassic/jurassic", "-db-migrations", "/jurassic/db/migrations"]
