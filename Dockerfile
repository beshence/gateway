FROM golang:1.26 AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /usr/local/bin/gateway .

FROM alpine:3.23.3

WORKDIR /usr/local/bin

RUN apk add --no-cache curl

COPY --from=builder /usr/local/bin/gateway /usr/local/bin/gateway

EXPOSE 27462

ENTRYPOINT ["/usr/local/bin/gateway"]


