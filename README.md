# Jurrasic Park API

Jurrasic Park API allows to keep track of different cages around the park and dinosaurs in the them.

## API specification

The OpenAPI specification for the API can be found in [`api/spec.yaml`](api/spec.yaml).

## API Reference

The generated API reference is available at https://jurassicpark.readme.io/reference

## Creating and initializing the database

Run the following statements to create a service account and a database for the API:

```sql
CREATE ROLE jurassic WITH LOGIN PASSWORD 'secret' NOSUPERUSER NOCREATEDB NOCREATEROLE;
CREATE DATABASE jurassic OWNER jurassic;
```

## Running the API

The address and port for the API can be set via `addr` flag or `JURASSIC_ADDR` environment variable. By default the API will listen on `:9001`.

The DB connection string should be passed via `JURASSIC_DB_CONN` environment variable.

To run the API with API key authentication enabled, the key should be set via `JURASSIC_API_KEY` environment variable.

To set the base URI for the API either pass it via `base-uri` flag or set `JURASSIC_BASE_URI` environment variable.

## Try it out

Add a new cage:

```bash
curl --request POST \
     --url http://localhost:9001/cages \
     --header 'accept: application/json' \
     --header 'content-type: application/json' \
     --data '{"capacity": 10, "status":"active"}'
  ```

List all cages:

```bash
curl --request GET \
     --url http://localhost:9001/cages \
     --header 'accept: application/json'
```

See more examples in the [documentation](https://jurassicpark.readme.io/reference).

## Local development

### Dependencies

To install development dependencies run:

```bash
make dev
```

### DB Migrations

All changes to the DB schema have to be done via migrations. To create a new migration run:

```bash
migrate create -ext sql -dir db/migrations -seq <migration_name>
```

Manually migrate up:

```bash
migrate -source file://db/migrations -database "$JURASSIC_DB_CONN" up
```

Manually migrate down:

```bash
migrate -source file://db/migrations -database "$JURASSIC_DB_CONN" down
```

### Testing

To run unit tests:

```bash
make test-unit
```

To run integration tests:

In order to run integration tests you need to create a separate database (e.g. `jurassic_test`) and optionally a service account (e.g. `jurassic_test`). To pass a connection string for the test database you can use `JURASSIC_TEST_DB_CONN` environment variable.

To run integration tests:

```bash
make test-integration
```

### Running API locally

If you have a local PostgreSQL instance with provisioned service account and the database:

```bash
make run
```

Using docker-compose:

```bash
docker-compose up
```
