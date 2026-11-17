// qdis - An efficient in-memory cache in Go
// Copyright (C) 2026 - 2027 Yixiang Qiu. All rights reserved.
// Use of this source code is governed by GPL-3. See LICENSE.

// Package server implements the qdis server.
package server

import (
	"fmt"
	"log"
	"net"

	"github.com/unsigend/qdis/internal/resp"
)

// Server is the qdis server.
type Server struct {
	addr string
}

// New creates a new server.
func New(addr string) *Server {
	return &Server{
		addr: addr,
	}
}

// serve serves a single connection.
func serve(conn net.Conn) error {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close connection: %v", err)
		}
	}()

	r := resp.NewReader(conn)

	for {
		cmd, err := r.ReadCmd()
		if err != nil {
			return err
		}

		// TODO: Handle command in cmd tables. For now just echo the command.
		fmt.Println(cmd)
	}

}

// ListenAndServe listens on the TCP network address s.addr and serves requests
// to the server.
func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	defer func() {
		if err := listener.Close(); err != nil {
			log.Printf("failed to close listener: %v", err)
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		// Connection model is goroutine-based, one connection per goroutine.
		go func() {
			if err := serve(conn); err != nil {
				log.Printf("failed to serve connection: %v", err)
			}
		}()

	}
}
