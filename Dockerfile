FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/remote-web-terminal .

FROM alpine:3.19
RUN apk add --no-cache bash ca-certificates tzdata
WORKDIR /app
COPY --from=builder /bin/remote-web-terminal /usr/local/bin/remote-web-terminal
COPY web/ ./web/

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/remote-web-terminal"]
