FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/seed ./cmd/seed

FROM alpine:3.22
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app
RUN mkdir -p /app/uploads/products && chown -R app:app /app/uploads
COPY --from=build /out/api /usr/local/bin/api
COPY --from=build /out/migrate /usr/local/bin/migrate
COPY --from=build /out/seed /usr/local/bin/seed
COPY migrations ./migrations
COPY seeds ./seeds
ENV MIGRATIONS_DIR=/app/migrations
ENV SEEDS_DIR=/app/seeds
USER app
EXPOSE 8080
CMD ["/usr/local/bin/api"]
