FROM gcr.io/distroless/static

USER nonroot:nonroot

COPY --chown=nonroot:nonroot jurassic /jurassic/jurassic
COPY --chown=nonroot:nonroot db/migrations /jurassic/db/migrations

CMD ["/jurassic/jurassic", "-db-migrations", "/jurassic/db/migrations"]
