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
