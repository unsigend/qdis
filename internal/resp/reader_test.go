// qdis - An efficient in-memory cache in Go
// Copyright (C) 2026 - 2027 Yixiang Qiu. All rights reserved.
// Use of this source code is governed by GPL-3. See LICENSE.

package resp

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"testing"
)

func TestReadCmd(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    [][]byte
		wantErr error
	}{
		{
			name: "PING",
			in:   "*1\r\n$4\r\nPING\r\n",
			want: [][]byte{[]byte("PING")},
		},
		{
			name: "ECHO x",
			in:   "*2\r\n$4\r\nECHO\r\n$1\r\nx\r\n",
			want: [][]byte{[]byte("ECHO"), []byte("x")},
		},
		{
			name: "nil bulk arg",
			in:   "*2\r\n$3\r\nGET\r\n$-1\r\n",
			want: [][]byte{[]byte("GET"), nil},
		},
		{
			name: "empty bulk arg",
			in:   "*2\r\n$3\r\nSET\r\n$0\r\n\r\n",
			want: [][]byte{[]byte("SET"), {}},
		},
		{
			name: "bulk with CRLF",
			in:   "*2\r\n$3\r\nSET\r\n$4\r\na\r\nb\r\n",
			want: [][]byte{[]byte("SET"), []byte("a\r\nb")},
		},
		{
			name:    "null array",
			in:      "*-1\r\n",
			wantErr: ErrInvalidValue,
		},
		{
			name:    "empty array",
			in:      "*0\r\n",
			wantErr: ErrInvalidValue,
		},
		{
			name:    "not array",
			in:      "+PONG\r\n",
			wantErr: ErrInvalidValue,
		},
		{
			name:    "nested array",
			in:      "*1\r\n*1\r\n$4\r\nPING\r\n",
			wantErr: ErrInvalidValue,
		},
		{
			name:    "truncated",
			in:      "*1\r\n$4\r\nPI",
			wantErr: io.ErrUnexpectedEOF,
		},
		{
			name:    "empty input",
			in:      "",
			wantErr: io.EOF,
		},
		{
			name:    "missing CR",
			in:      "*1\n$4\r\nPING\r\n",
			wantErr: ErrBadFmt,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewReader(bytes.NewReader([]byte(tc.in))).ReadCmd()
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("error = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ReadCmd: %v", err)
			}
			if !argsEqual(got, tc.want) {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestReadCmdPipeline(t *testing.T) {
	in := "*1\r\n$4\r\nPING\r\n*2\r\n$4\r\nECHO\r\n$1\r\nx\r\n"
	r := NewReader(bytes.NewReader([]byte(in)))

	a, err := r.ReadCmd()
	if err != nil {
		t.Fatal(err)
	}
	if !argsEqual(a, [][]byte{[]byte("PING")}) {
		t.Fatalf("first = %q", a)
	}

	b, err := r.ReadCmd()
	if err != nil {
		t.Fatal(err)
	}
	if !argsEqual(b, [][]byte{[]byte("ECHO"), []byte("x")}) {
		t.Fatalf("second = %q", b)
	}
}

func TestReadBulk(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    []byte
		wantNil bool
		wantErr error
	}{
		{name: "hello", in: "$5\r\nhello\r\n", want: []byte("hello")},
		{name: "empty", in: "$0\r\n\r\n", want: []byte{}},
		{name: "null", in: "$-1\r\n", wantNil: true},
		{name: "crlf in body", in: "$4\r\na\r\nb\r\n", want: []byte("a\r\nb")},
		{name: "negative", in: "$-2\r\n", wantErr: ErrInvalidValue},
		{name: "too large", in: "$" + strconv.Itoa(maxBulk+1) + "\r\n", wantErr: ErrTooLargeBulk},
		{name: "truncated body", in: "$5\r\nhel", wantErr: io.ErrUnexpectedEOF},
		{name: "bad prefix", in: "+OK\r\n", wantErr: ErrInvalidValue},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewReader(bytes.NewReader([]byte(tc.in))).ReadBulk()
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("error = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ReadBulk: %v", err)
			}
			if tc.wantNil {
				if got != nil {
					t.Fatalf("got %q, want nil", got)
				}
				return
			}
			if !bytes.Equal(got, tc.want) {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestReadLineTypes(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		got, err := NewReader(bytes.NewReader([]byte("+PONG\r\n"))).ReadString()
		if err != nil || got != "PONG" {
			t.Fatalf("got %q %v", got, err)
		}
	})
	t.Run("error", func(t *testing.T) {
		got, err := NewReader(bytes.NewReader([]byte("-ERR foo\r\n"))).ReadError()
		if err != nil || got != "ERR foo" {
			t.Fatalf("got %q %v", got, err)
		}
	})
	t.Run("int", func(t *testing.T) {
		got, err := NewReader(bytes.NewReader([]byte(":-1\r\n"))).ReadInt()
		if err != nil || got != -1 {
			t.Fatalf("got %d %v", got, err)
		}
	})
	t.Run("int junk", func(t *testing.T) {
		_, err := NewReader(bytes.NewReader([]byte(":1x\r\n"))).ReadInt()
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func FuzzReadCmd(f *testing.F) {
	f.Add([]byte("*1\r\n$4\r\nPING\r\n"))
	f.Add([]byte("*2\r\n$4\r\nECHO\r\n$1\r\nx\r\n"))
	f.Add([]byte("*1\r\n$4\r\nPI"))
	f.Add([]byte("*-1\r\n"))
	f.Add([]byte{})
	f.Fuzz(func(_ *testing.T, data []byte) {
		_, err := NewReader(bytes.NewReader(data)).ReadCmd()
		if err != nil {
			return
		}
	})
}

func argsEqual(a, b [][]byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if (a[i] == nil) != (b[i] == nil) {
			return false
		}
		if !bytes.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}
