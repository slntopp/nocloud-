FROM golang:1.17-alpine AS builder

RUN apk add upx

WORKDIR /go/src/github.com/slntopp/nocloud-tunnel-mesh
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" ./cmd/tunnelserver
RUN upx ./tunnelserver

FROM scratch
WORKDIR /
COPY --from=builder  /go/src/github.com/slntopp/nocloud-tunnel-mesh/tunnelserver /tunnelserver

#COPY cert cert

EXPOSE 8000
LABEL org.opencontainers.image.source https://github.com/slntopp/nocloud-tunnel-mesh

ENTRYPOINT ["/tunnelserver"]
