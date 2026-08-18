FROM golang:1.25-alpine@sha256:3eb6c2b3db8d55e38537302edb510b4417f8a115efbd5906d131ceba9468e29a AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/seed ./cmd/seed
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/media-reconcile ./cmd/media-reconcile
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/media-init ./cmd/media-init

# Test-only stage. The Go race detector requires CGO and a C toolchain; no
# compiler or CGO-linked binary is copied into production runtime images.
FROM build AS validation
RUN apk add --no-cache gcc musl-dev

FROM alpine:3.22@sha256:14358309a308569c32bdc37e2e0e9694be33a9d99e68afb0f5ff33cc1f695dce AS runtime
RUN apk add --no-cache ca-certificates && addgroup -S app && adduser -S app -G app
WORKDIR /app
USER app

FROM runtime AS api
USER root
RUN mkdir -p /app/uploads/products && chown -R app:app /app/uploads
COPY --from=build /out/api /usr/local/bin/api
USER app
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --retries=3 CMD wget -qO- http://localhost:8080/api/v1/health/ready || exit 1
CMD ["/usr/local/bin/api"]

FROM runtime AS worker
COPY --from=build /out/worker /usr/local/bin/worker
CMD ["/usr/local/bin/worker"]

FROM runtime AS migrate
COPY --from=build /out/migrate /usr/local/bin/migrate
COPY migrations /app/migrations
ENV MIGRATIONS_DIR=/app/migrations
CMD ["/usr/local/bin/migrate"]

FROM runtime AS seed
COPY --from=build /out/seed /usr/local/bin/seed
COPY seeds /app/seeds
ENV SEEDS_DIR=/app/seeds
CMD ["/usr/local/bin/seed"]

FROM runtime AS media-reconcile
COPY --from=build /out/media-reconcile /usr/local/bin/media-reconcile
CMD ["/usr/local/bin/media-reconcile"]

FROM runtime AS media-init
COPY --from=build /out/media-init /usr/local/bin/media-init
CMD ["/usr/local/bin/media-init"]
