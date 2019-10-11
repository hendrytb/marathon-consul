FROM golang:1.13.1-alpine as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o mesos-consul main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/mesos-consul .
