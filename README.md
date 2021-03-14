# **kvstore**

**kvstore** is a very simple redis-like (key-value pair) in-memory database server implemented in **Go** & hashtables

![kvstore in action](.github/kvstore.gif)

# Please Note

This project is not targeted for production use. The purpose of this project? Mostly just for fun and learning.

Want to expand this project? open up an [issue](https://github.com/HotPotatoC/kvstore/issues/new) or you can contact me on juandotulung@gmail.com

# Installation

You can install the binaries in the [releases](https://github.com/HotPotatoC/kvstore/releases) tab. Alternatively, to get the latest version of kvstore run:

```
❯ go get -u github.com/HotPotatoC/kvstore
```

# Getting started

Running the kvstore server using the `kvstore_server` command

```sh
❯ kvstore_server

2021-03-11T17:08:23.104+0700    KVStore is starting...
2021-03-11T17:08:23.106+0700    version=v1.0.0 build=6ccb99fc20a525ceb8ca384bd2b3967337661874 pid=1
2021-03-11T17:08:23.106+0700    starting tcp server...

         _               _
        | |             | |
        | | ____   _____| |_ ___  _ __ ___
        | |/ /\ \ / / __| __/ _ \| '__/ _ \
        |   <  \ V /\__ \ || (_) | | |  __/
        |_|\_\  \_/ |___/\__\___/|_|  \___|

        Started KVStore v1.0.0 server
          Port: 7275
          PID: 1

2021-03-11T17:08:23.107+0700    Ready to accept connections.
```

To interact with the server, on another terminal run the `kvstore_cli` command

```sh
❯ kvstore_cli

127.0.0.1:7275> info
0.000704s
{
  "os": "linux",
  "os_arch": "amd64",
  "go_version": "go1.16.2",
  "process_id": 27120,
  "tcp_host": "0.0.0.0",
  "tcp_port": 7275,
  "server_uptime": 74754094200,
  "server_uptime_human": "1m14.7540944s",
  "connected_clients": 1,
  "total_connections_count": 1,
  "memory_usage": 510888,
  "memory_usage_human": "498.9 kB",
  "memory_total_alloc": 510888
}


127.0.0.1:7275>
```

# Command Table

| Command (Case insensitive) | Description                                                                                      |
| -------------------------- | ------------------------------------------------------------------------------------------------ |
| SET [key] [value]          | Inserts a new entry into the database                                                            |
| GET [key]                  | Returns the data in the database with the matching key                                           |
| DEL [key]                  | Remove an entry in the database with the matching key                                            |
| LIST                       | Displays all the saved data in the database with the format `[key] -> [value]`                   |
| KEYS                       | Displays all the saved keys in the database                                                      |
| INFO                       | Displays the current stats of the server (OS, mem usage, total connections, etc.) in json format |

# Todo

- Wildcard pattern matching
- Open addressing implementation(?)
- ...

# Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

# License

[MIT](https://choosealicense.com/licenses/mit/)
