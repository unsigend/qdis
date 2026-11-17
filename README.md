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
make vet                 - run go vet
make check               - format, vet, and lint
make clean               - remove the build directory
make line                - count implementation lines
make version             - print the project version
make help                - show this help message
```

Build output goes to `build/`.

## Contribution and Copyright

Contributions are welcome. Copyright (C) 2026–2027 Yixiang Qiu. qdis is released under the GNU GPL v3. See `LICENSE`.
