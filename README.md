# Ethereum 2.0 Network Stats and Monitoring - CLI Client

> This is an intial POC release of the eth2stats network monitoring suite
> 
> Currently it supports just Prysm and their public testnet.
> More to come soon.

## Supported clients and protocols:
- [ ] Prysm
  - [x] GRPC
  - [ ] HTTP
- [ ] ...
  
## Current live deployments:

- [https://sapphire.eth2stats.net/](https://sapphire.eth2stats.net/) - [Prysm Sapphire Testnet](https://prylabs.net/participate)

## Getting Started (Prysm Sapphire Testnet)

The first thing you should do is get a beacon chain client node running and connected to said beacon chain by joining the [Prysm Sapphire Testnet](https://prylabs.net/participate).

Then you can run the following Docker command to start sending stats to [eth2stats](https://sapphire.eth2stats.net).  
**Please update your --eth2stats.node-name before starting the cli tool.**
```shell script
docker run -d --name eth2stats-client --network="host" \
    alethio/eth2stats-client:latest \
    run --v \
    --eth2stats.node-name="YourNode" \
    --eth2stats.addr="grpc.sapphire.eth2stats.net:443" --eth2stats.tls=true \
    --beacon.type="prysm" --beacon.addr="localhost:4000"
```

You should now be able to see your node and it's stats on [eth2stats](https://sapphire.eth2stats.net).

