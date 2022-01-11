# Gocore Server Prometheus Exporter [![](https://images.microbadger.com/badges/image/core-coin/xcbexporter.svg)](https://microbadger.com/images/core-coin/xcbexporter)
Monitor your Core-coin Gocore server with Prometheus and Grafana. Checkout the [Grafana Dashboard](https://grafana.com/dashboards/6976) to implement a beautiful gocore server monitor for your own server, or you can just import Dashboard ID: `6976` once you have xcbexporter up and running.

<p align="center"><img width="90%" src="https://img.cjx.io/xcbexporter-grafana.png"></p>

## Docker
Run this Prometheus Exporter in a [Docker container](https://hub.docker.com/r/core-coin/xcbexporter/builds/)! Include your Gocore server endpoint as `GOCORE` environment variable.
```bash
docker run -it -d -p 9090:9090 \
  -e "GOCORE="http://mygocoreserverhere.com:8545" \
  core-coin/xcbexporter
```

## Features
- Current and Average Energy Price
- Total amount of ERC20 Token Transfers
- Total amount of XCB transactions
- Watch balance on specific addresses
- Pending Transaction count

## Environment Variables
You can add the environment variable `ADDRESSES` with a comma delimited list of core-coin addresses.

- `GOCORE` = `http://xcb.mygocoreserver.com:8545` Core-coin node endpoint
- `ADDRESSES` = `0x867fFB5a3871b500f65BdFafe0136f9667Deae06,0xF008E2c7A7F16ac706C2E0EBD3F015D442016420`
- `DELAY` = `500` millisecond delay between requests

## Prometheus Response
```
gocore_block 7042028
gocore_seconds_last_block 0.50
gocore_block_transactions 48
gocore_block_value 59.48321713266354
gocore_block_energy_used 1243863
gocore_block_energy_limit 8000000
gocore_block_nonce 7516583072599285197
gocore_block_difficulty 2606288773636567
gocore_block_uncles 0
gocore_block_size_bytes 6680
gocore_energy_price 2000000000
gocore_pending_transactions 136
gocore_network_id 1
gocore_contracts_created 0
gocore_token_transfers 10
gocore_xcb_transfers 35
gocore_load_time 0.5302
gocore_address_balance{address="0x867fFB5a3871b500f65BdFafe0136f9667Deae06"} 86.99212193
gocore_address_nonce{address="0x867fFB5a3871b500f65BdFafe0136f9667Deae06"} 1
gocore_address_balance{address="0xF008E2c7A7F16ac706C2E0EBD3F015D442016420"} 0.1605609476
gocore_address_nonce{address="0xF008E2c7A7F16ac706C2E0EBD3F015D442016420"} 95623
```
