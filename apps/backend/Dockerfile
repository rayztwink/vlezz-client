FROM golang:1.24-alpine AS build
WORKDIR /src
COPY apps/backend/go.mod apps/backend/go.sum* ./
RUN go mod download
COPY apps/backend/ ./
RUN go build -o /out/rayflowd ./cmd/rayflowd

FROM alpine:3.21
RUN adduser -D -H rayflow
WORKDIR /app
COPY --from=build /out/rayflowd /usr/local/bin/rayflowd
COPY apps/backend/migrations ./migrations
RUN mkdir -p data configs logs presets && chown -R rayflow:rayflow /app
USER rayflow
EXPOSE 8787
CMD ["rayflowd"]

