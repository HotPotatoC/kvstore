# **kvstore**

**kvstore** is a very simple redis-like (key-value pair) in-memory database server implemented in **Go** & hashtables

![kvstore in action](.github/kvstore.gif)

# Please Note

This project is not targeted for production use.

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
2021-03-11T17:08:23.106+0700    version=v0.4.2 build=c0a2dc7711ae0d61870f6e248f1205d23c31d808 pid=1182
2021-03-11T17:08:23.106+0700    starting tcp server...

         _               _
        | |             | |
        | | ____   _____| |_ ___  _ __ ___
        | |/ /\ \ / / __| __/ _ \| '__/ _ \
        |   <  \ V /\__ \ || (_) | | |  __/
        |_|\_\  \_/ |___/\__\___/|_|  \___|

        Started KVStore v0.4.2 server
          Port: 7275
          PID: 1182

2021-03-11T17:08:23.107+0700    Ready to accept connections.
```

To interact with the server, on another terminal run the `kvstore_cli` command

```sh
❯ kvstore_cli

127.0.0.1:7275>
```

# Command Table

| Command (Case insensitive) | Description                                                                    |
| -------------------------- | ------------------------------------------------------------------------------ |
| SET [key] [value]          | Inserts a new entry into the database                                          |
| GET [key]                  | Returns the data in the database with the matching key                         |
| DEL [key]                  | Remove an entry in the database with the matching key                          |
| LIST                       | Displays all the saved data in the database with the format `[key] -> [value]` |
| KEYS                       | Displays all the saved keys in the database                                    |

# Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

# License

[MIT](https://choosealicense.com/licenses/mit/)
