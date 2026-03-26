FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
WORKDIR /app/cmd/worker
RUN go build -o /go/bin/app

FROM alpine:latest
COPY --from=builder /go/bin/app /app
EXPOSE 8080
CMD ["/app"]
