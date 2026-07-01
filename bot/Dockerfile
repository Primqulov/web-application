FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod ./
RUN go mod download || true
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/bot ./cmd/bot
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/feedbackbot ./cmd/feedbackbot

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/bot /app/bot
COPY --from=build /out/feedbackbot /app/feedbackbot
ENTRYPOINT ["/app/bot"]
