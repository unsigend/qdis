// qdis - An efficient in-memory cache in Go
// Copyright (C) 2026 - 2027 Yixiang Qiu. All rights reserved.
// Use of this source code is governed by GPL-3. See LICENSE.

package server

import (
	"io"
	"net"
	"testing"
	"time"
)

func TestServePipeline(t *testing.T) {
	c1, c2 := net.Pipe()
	t.Cleanup(func() {
		if err := c1.Close(); err != nil {
			t.Errorf("close c1: %v", err)
		}
		if err := c2.Close(); err != nil {
			t.Errorf("close c2: %v", err)
		}
	})
	if err := c1.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := c2.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatal(err)
	}

	go func() {
		if err := serve(c2); err != nil {
			return
		}
	}()

	want := "+PONG\r\n$1\r\nx\r\n"
	gotc := make(chan []byte, 1)
	errc := make(chan error, 1)
	go func() {
		buf := make([]byte, len(want))
		_, err := io.ReadFull(c1, buf)
		if err != nil {
			errc <- err
			return
		}
		gotc <- buf
	}()

	in := "*1\r\n$4\r\nPING\r\n*2\r\n$4\r\nECHO\r\n$1\r\nx\r\n"
	if _, err := c1.Write([]byte(in)); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-errc:
		t.Fatalf("read: %v", err)
	case got := <-gotc:
		if string(got) != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout")
	}
}
