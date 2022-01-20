FROM golang:1.17-alpine AS builder

RUN apk add upx

WORKDIR /go/src/github.com/slntopp/nocloud-tunnel-mesh
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" ./cmd/tunnelclient
RUN upx ./tunnelclient

FROM alpine:3.13
WORKDIR /
COPY --from=builder  /go/src/github.com/slntopp/nocloud-tunnel-mesh/tunnelclient /tunnelclient
# COPY /go/src/github.com/slntopp/nocloud-tunnel-mesh/tunnelclient /tunnelclient

RUN     apk update \
    &&  apk add bash bind-tools busybox-extras curl \
                iproute2 iputils jq mtr \
                net-tools openssl \
                perl-net-telnet procps tcpdump tcptraceroute 


COPY hack/client_debug.sh /start

EXPOSE 8000
LABEL org.opencontainers.image.source https://github.com/slntopp/nocloud-tunnel-mesh

ENTRYPOINT ["/start"]
