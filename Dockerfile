# build for amd64
FROM golang:1.13.5 AS build

ENV ROOT "/tmp/eth2stats-client"

WORKDIR $ROOT

ADD go.mod go.mod
ADD go.sum go.sum

ADD . .

RUN make build

# create application container for amd64
FROM ubuntu

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /

COPY --from=build /tmp/eth2stats-client/eth2stats-client .

ENTRYPOINT ["/eth2stats-client"]
