# ---- Build Stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o arr-proxy ./cmd

# ---- Final Stage ----
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/arr-proxy .
COPY --from=builder /app/config /config

ENV APP_CONFIG_DIR=/config

EXPOSE 8443

CMD ["./arr-proxy"]
