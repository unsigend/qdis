// qdis - An efficient in-memory cache in Go
// Copyright (C) 2026 - 2027 Yixiang Qiu. All rights reserved.
// Use of this source code is governed by GPL-3. See LICENSE.

// Package server implements the qdis server.
package server

import (
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/unsigend/qdis/internal/resp"
	"github.com/unsigend/qdis/internal/store"
)

// Server is the qdis server.
type Server struct {
	addr  string
	store *store.Store
}

// NewServer creates a new server.
func NewServer(addr string, store *store.Store) *Server {
	return &Server{
		addr:  addr,
		store: store,
	}
}

// serve serves a single connection.
func serve(conn net.Conn, store *store.Store) error {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close connection: %v", err)
		}
	}()

	r := resp.NewReader(conn)
	w := resp.NewWriter(conn)

	for {
		cmd, err := r.ReadCmd()
		if err != nil {
			return err
		}

		command, err := NewCommandRaw(cmd)
		if err != nil {
			return err
		}

		if err := command.RunCommand(w, store); err != nil {
			if errors.Is(err, ErrInvalidCommand) { // unknown command
				errstr := fmt.Sprintf("ERR unknown command '%s'", command.Name)
				if werr := w.WriteError(errstr); werr != nil {
					return werr
				}
			} else if errors.Is(err, ErrQuit) { // quit command
				if ferr := w.Flush(); ferr != nil {
					return ferr
				}
				break
			} else {
				return err
			}
		}

		if err := w.Flush(); err != nil {
			return err
		}
	}

	return nil
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
			if err := serve(conn, s.store); err != nil {
				log.Printf("failed to serve connection: %v", err)
			}
		}()

	}
}
