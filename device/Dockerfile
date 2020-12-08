FROM golang:1.15-alpine as builder
WORKDIR /opt
COPY . .
RUN go mod download

WORKDIR /opt/device
RUN CGO_ENABLED=0 go build -o /opt/bin .

FROM alpine:3.12
WORKDIR /opt
RUN apk update && apk add bash --no-cache
COPY --from=builder /opt/bin ./bin
