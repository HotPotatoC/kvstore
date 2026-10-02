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

- `SET key value [NX | XX] [EX seconds | PX milliseconds]`
- `GET key`
- `DEL key [key ...]` (literal keys, including `*`)
- `EXPIRE key seconds`, `PEXPIRE key milliseconds`
- `TTL key`, `PTTL key`
- `KEYS pattern`
- `SCAN cursor [MATCH pattern] [COUNT count]`
- `INFO [server | clients | memory | stats | persistence | keyspace | default | all]`
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

KEYS, SCAN, INFO, FLUSHALL, and CLIENT LIST/KILL run on `server.workers` workers
(default four), with `server.worker_queue` queued commands (default 16). One slow
command runs per connection; later commands wait for its response. Saturation
returns `ERR server overloaded` for that command. Responses exceeding the output
limit return an error or close the connection.

`database.maxmemory` defaults to 256 MiB: key/value bytes plus an estimate of
item and index metadata. Writes that exceed this budget return `OOM`; existing
values and TTLs remain intact. Deletes, expiry, and smaller replacements free
budget. There is no eviction. This is not a process RSS cap: Go runtime,
allocator overhead, snapshot decoding, and connection buffers need extra memory.
`server.maxclients` defaults to 1,000 concurrent connections. Set either limit
to `0` to disable it. Snapshots exceeding the data budget fail to load.

Snapshots save on clean shutdown using a synced temporary file, atomic rename,
and directory sync. A failed replacement before rename preserves the previous
snapshot; a sync error after rename is reported. Crashes still lose changes
since the last snapshot. SET expiry must be positive; EXPIRE/PEXPIRE with zero
or negative expiry deletes immediately. Positive expiry values exceeding Go's duration
range (about 292 years) are rejected. EXPIRE condition flags are unsupported.

`INFO` returns RESP bulk text with Server, Clients, Memory, Stats, Persistence,
and Keyspace sections. No argument, `default`, and `all` include all six sections;
section names are case insensitive, and unknown sections return empty text.
`used_memory` is the stored-data budget accounting described above, not Go heap
or process RSS. Keyspace counts read shard index sizes without scanning keys;
expired entries awaiting cleanup remain counted. Values can reflect different
instants during concurrent changes.

`total_connections_received` counts accepted connections; `rejected_connections`
counts maxclients refusals. `total_commands_processed` counts dispatched commands,
including command errors and INFO itself, across event loops and workers.
`total_error_replies` counts generated errors, including protocol and overload
errors. `rejected_requests` counts OOM, worker overload, protocol failures, and
input/output limit rejections. Queue rejections are not processed commands.
`oom_errors` and `overload_errors` report those rejection causes separately.
Persistence fields report snapshot saves and clears in the current process:
`snapshot_last_save_status` is `never`, `ok`, or `err`; failures retain the last
successful save's Unix timestamp (zero until a successful save). INFO respects
the configured response limit and reports save status without waiting for disk I/O.

Use `SCAN 0` to begin iteration, then pass each returned cursor until it is `0`.
Empty pages with nonzero cursors are valid. `COUNT` defaults to 10 and is capped
at 1,024 inspected slots per call, including deleted slots and nonmatches.
Pages also obey the response limit and a soft 1 ms work budget. `MATCH` supports
Redis byte glob patterns (`*`, `?`, character classes, and escapes); pathological
matches exceeding one million matching steps return an error. `TYPE` is unsupported.
Keys present throughout an iteration are returned; concurrent additions and
deletions may appear or be omitted. Cursors are not retained across restarts.
`redis-cli --scan --pattern 'prefix:*'` works. KEYS still performs a full scan.

Expiry cleanup runs every 10 ms, checks at most 1,024 TTL entries per cycle,
and checks a soft 1 ms time budget between batches of at most 32 entries.
Expired keys disappear on access immediately; background reclamation may lag
during large expiry bursts. Deleted scan slots remain charged to the memory
budget until reused or their shard becomes empty. Spare array capacity, like
allocator slack, remains outside the estimate.

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
