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
- `VALUES`
- `PING`
- `FLUSHALL`
- `CLIENT [ID | INFO | LIST | KILL <id | addr | user> <value> | GETNAME | SETNAME <name>]`

## NOTE

This project is not targeted for production use. This is only a proof of concept

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[MIT](https://choosealicense.com/licenses/mit/)

## Support

<a href="https://www.buymeacoffee.com/hotpotato" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: 41px !important;width: 174px !important;box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;-webkit-box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;" ></a>
