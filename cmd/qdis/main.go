// qdis - An efficient in-memory cache in Go
// Copyright (C) 2026 - 2027 Yixiang Qiu. All rights reserved.
// Use of this source code is governed by GPL-3. See LICENSE.

// Package main is the main package for the qdis server.
package main

import (
	"flag"
	"fmt"
)

const (
	defaultAddr = "127.0.0.1:6379"
)

func main() {
	addr := flag.String("addr", defaultAddr, "listen address")

	flag.Parse()

	fmt.Println("Listening on", *addr)
}
