FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod ./
RUN go mod download || true
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/seed ./seed

FROM alpine:3.19
RUN apk add --no-cache ca-certificates && mkdir -p /data/avatars
WORKDIR /app
COPY --from=build /out/api /app/api
COPY --from=build /out/seed /app/seed
EXPOSE 8080
ENTRYPOINT ["/app/api"]
