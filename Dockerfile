FROM golang:1.24.2 AS base

# Deps Stage
FROM base AS deps-stage

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

# Files Copy Stage
FROM deps-stage AS files-copy-stage

WORKDIR /app

COPY go.mod go.sum ./
COPY /pkg ./pkg
COPY /internal ./internal
COPY /cmd ./cmd

# Build Stage
FROM files-copy-stage AS build-stage

WORKDIR /app/cmd/url-shortener

RUN go build -o /url-shortener-api .

# Environment 
FROM alpine:3.20.1 AS environment-stage

RUN apk add --no-cache gcompat=1.1.0-r4

WORKDIR /app

COPY --from=build-stage /url-shortener-api ./url-shortener-api

CMD [ "./url-shortener-api" ]