FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY go.mod .
COPY services/go/ services/go/
COPY gateway/ gateway/

ARG SERVICE_PATH
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app ${SERVICE_PATH}

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /app /app
ENTRYPOINT ["/app"]
