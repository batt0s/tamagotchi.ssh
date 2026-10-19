# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /out/tamagotchi-ssh ./cmd/pet

# Runtime stage
FROM alpine:3.22
RUN adduser -D -u 10001 battos
WORKDIR /app
COPY --from=build /out/tamagotchi-ssh /app/tamagotchi-ssh
COPY balance.yaml /app/balance.yaml
RUN mkdir -p /app/.ssh /app/data && chown -R battos:battos /app
USER battos
EXPOSE 2222
ENTRYPOINT ["/app/tamagotchi-ssh"]
