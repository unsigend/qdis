# qdis

qdis is an efficient in-memory cache written in Go. It speaks the Redis protocol and is intended for use with `redis-cli` and `redis-benchmark`.

Current version: `v0.1.0`

For copyright information, see the file `LICENSE` in this directory.

## Build

```
make all                 - build the server binary
make build               - build the server binary
make test                - run the test suite
make race                - run tests with the race detector
make fmt                 - format Go sources
make fmtcheck            - fail if sources are unformatted
make vet                 - run go vet
make check               - format, vet, and lint
make ci                  - fmtcheck, vet, lint, and test
make clean               - remove the build directory
make line                - count implementation lines
make version             - print the project version
make help                - show this help message
```

Build output goes to `build/`.

## Commands

Syntax matches Redis. Only listed forms are implemented.

### Connection

**`PING`** `[message]`  
Return: simple string `PONG`, or bulk string if a message is given.

**`ECHO`** `message`  
Return: bulk string reply of `message`.

**`QUIT`**  
Return: simple string `OK`. The connection is then closed.

### Generic

**`DEL`** `key [key ...]`  
Remove the given keys. Missing keys are ignored.  
Return: integer, number of keys that were removed.

**`EXISTS`** `key [key ...]`  
Count how many of the given keys exist. Repeated names are counted again.  
Return: integer, number of existing keys among the arguments.

**`TYPE`** `key`  
Return the type of the value stored at key.  
Return: simple string `string`, or `none` if the key does not exist.

**`DBSIZE`**  
Return: integer, number of keys in the current database.

**`FLUSHDB`**  
Remove all keys from the current database.  
Return: simple string `OK`.

### String

**`SET`** `key value`  
Set key to hold the string value. Overwrites any existing value.  
Return: simple string `OK`.

**`GET`** `key`  
Get the string value of key.  
Return: bulk string, or Nil when key does not exist.



## Contribution and Copyright

Contributions are welcome. Copyright (C) 2026–2027 Yixiang Qiu. qdis is released under the GNU GPL v3. See `LICENSE`.
