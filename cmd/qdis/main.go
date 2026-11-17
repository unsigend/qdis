// qdis - An efficient in-memory cache in Go
// Copyright (C) 2026 - 2027 Yixiang Qiu. All rights reserved.
// Use of this source code is governed by GPL-3. See LICENSE.

// Package main is the main package for the qdis server.
package main

import (
	"flag"
	"log"

	"github.com/unsigend/qdis/internal/server"
)

const (
	defaultAddr = "127.0.0.1:6379"
)

func main() {
	addr := flag.String("addr", defaultAddr, "listen address")

	flag.Parse()

	s := server.New(*addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("failed to listen and serve: %v", err)
	}
}
