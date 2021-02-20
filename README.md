# Todo

- [x] TCP Server
- [x] CLI
- [x] Command & Args parsing
- [ ] Command implementation (`set`, `del`, `get` and `list`)
- [ ] Hash Table chaining to avoid collision
- [ ] TCP connection pooling (?)

# Setting up (WIP)

1. Clone the repo
2. Run the kvstore server

```sh
❯ go run cmd/kvstore_server/main.go

2021-02-20T11:36:19.320+0700    KVStore is starting...
2021-02-20T11:36:19.320+0700    starting tcp server... pid=9325

         _               _
        | |             | |
        | | ____   _____| |_ ___  _ __ ___
        | |/ /\ \ / / __| __/ _ \| '__/ _ \
        |   <  \ V /\__ \ || (_) | | |  __/
        |_|\_\  \_/ |___/\__\___/|_|  \___|

        Started KVStore server
          Port: 7275
          PID: 9325

2021-02-20T11:36:19.321+0700    Ready to accept connections.
```

3. Run the kvstore CLI on another terminal

```sh
❯ go run cmd/kvstore_cli/main.go

127.0.0.1:7275>
```

4. Run simple commands like `set` and `get`

```sh
127.0.0.1:7275> set user::1 {"name":"juan","email":"juandotulung@gmail.com"}

127.0.0.1:7275> get user::1
{"name":"juan","email":"juandotulung@gmail.com"}

127.0.0.1:7275> get user::2
<nil>
```
