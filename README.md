# **kvstore**

An experimental key-value database server that is compatible with the redis **RESP** protocol.

## Getting started

Simply run the following command to start the server:

```bash
go run cmd/kvstore-server/main.go
```

To connect to the server, currently the `kvstore-cli` is yet to be implemented. So for now, you can use the `redis-cli` command to connect to the server.

```bash
redis-cli -p 7275 # Default kvstore server port is 7275
```

Current available commands are:

- `SET key value`
- `GET key`
- `DEL key`
- `KEYS pattern`
- `PING`
- `ECHO value`
- `FLUSHALL`
- `CLIENT [ID | INFO | LIST | KILL <id | addr | user> <value> | GETNAME | SETNAME <name>]`

The server defaults to four event loops. Each traffic callback yields after
64 commands or after response data reaches 64 KiB. One large response can exceed
that callback budget; the pending-output limit remains a hard cap. `server.loops`,
`server.command_budget`, and `server.output_budget` configure these limits.

`server.max_pending_input` and `server.max_pending_output` default to 16 MiB per
connection. Connections exceeding either pending-byte limit are closed. Commands
accept at most 1,024 bulk arguments, each at most 8 MiB, and a 16 MiB total frame.
Incomplete frames remain buffered until more input arrives.

KEYS, pattern DEL, FLUSHALL, and CLIENT LIST/KILL run on `server.workers` workers
(default four), with `server.worker_queue` queued commands (default 16). One slow
command runs per connection; later commands wait for its response. Saturation
returns `ERR server overloaded` for that command. Responses exceeding the output
limit return an error or close the connection.

## Benchmarks

Throughput in requests/second, median of three runs on an Apple M4 Pro
(14 cores, 24 GiB), localhost, 10,000 keys. Measured kvstore `8f14438` with four
event loops against Valkey 9.0.0; persistence disabled during measurement.

| Workload | kvstore | Valkey 9.0.0, default | Valkey 9.0.0, 4 I/O threads |
|---|---:|---:|---:|
| SET, 32 B, no pipeline | 126.6k | 145.6k | 145.8k |
| GET, 32 B, no pipeline | 130.7k | 142.9k | 148.1k |
| SET, 32 B, pipeline 16 | 2.664M | 1.176M | 2.498M |
| GET, 32 B, pipeline 16 | 2.663M | 1.481M | 2.497M |
| SET, 1 KiB, pipeline 16 | 1.998M | 1.000M | 1.998M |
| GET, 1 KiB, pipeline 16 | 1.998M | 1.142M | 1.998M |

Used `valkey-benchmark` 9.0.0: no pipeline = 50 clients, one benchmark thread,
300k requests; pipeline 16 = 200 clients, four benchmark threads, 10M requests
for 32 B or 2M for 1 KiB. Small differences overlap run-to-run variation.

## To Do

- [x] Pipelining commands
- [ ] AOF
- [ ] ACL
- [ ] Clustering
- [ ] Implement `kvstore-cli`

## NOTE

This project is not targeted for production use. This is only a proof of concept

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[MIT](https://choosealicense.com/licenses/mit/)

## Support

<a href="https://www.buymeacoffee.com/hotpotato" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: 41px !important;width: 174px !important;box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;-webkit-box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;" ></a>
