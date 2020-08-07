FROM golang:alpine as build

RUN apk add --no-cache ca-certificates build-base

WORKDIR /build

ADD go.mod go.mod
ADD go.sum go.sum

ADD . .

RUN make build

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt \
     /etc/ssl/certs/ca-certificates.crt

COPY --from=build /build/eth2stats-client /eth2stats-client

WORKDIR /

ENTRYPOINT ["/eth2stats-client"]
