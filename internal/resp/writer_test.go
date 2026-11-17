// qdis - An efficient in-memory cache in Go
// Copyright (C) 2026 - 2027 Yixiang Qiu. All rights reserved.
// Use of this source code is governed by GPL-3. See LICENSE.

package resp

import (
	"bytes"
	"errors"
	"testing"
)

func TestWriter(t *testing.T) {
	tests := []struct {
		name    string
		write   func(*Writer) error
		want    string
		wantErr error
	}{
		{
			name:  "simple PONG",
			write: func(w *Writer) error { return w.WriteString("PONG") },
			want:  "+PONG\r\n",
		},
		{
			name:  "simple empty",
			write: func(w *Writer) error { return w.WriteString("") },
			want:  "+\r\n",
		},
		{
			name:    "simple rejects CR",
			write:   func(w *Writer) error { return w.WriteString("PO\rNG") },
			wantErr: ErrInvalidValue,
		},
		{
			name:    "simple rejects LF",
			write:   func(w *Writer) error { return w.WriteString("PO\nNG") },
			wantErr: ErrInvalidValue,
		},
		{
			name:  "error",
			write: func(w *Writer) error { return w.WriteError("ERR unknown command") },
			want:  "-ERR unknown command\r\n",
		},
		{
			name:    "error rejects CR",
			write:   func(w *Writer) error { return w.WriteError("ERR \r") },
			wantErr: ErrInvalidValue,
		},
		{
			name:  "int zero",
			write: func(w *Writer) error { return w.WriteInt(0) },
			want:  ":0\r\n",
		},
		{
			name:  "int negative",
			write: func(w *Writer) error { return w.WriteInt(-1) },
			want:  ":-1\r\n",
		},
		{
			name:  "bulk hello",
			write: func(w *Writer) error { return w.WriteBulk([]byte("hello")) },
			want:  "$5\r\nhello\r\n",
		},
		{
			name:  "bulk empty",
			write: func(w *Writer) error { return w.WriteBulk([]byte{}) },
			want:  "$0\r\n\r\n",
		},
		{
			name:  "bulk null",
			write: func(w *Writer) error { return w.WriteBulk(nil) },
			want:  "$-1\r\n",
		},
		{
			name:  "bulk with CRLF",
			write: func(w *Writer) error { return w.WriteBulk([]byte("a\r\nb")) },
			want:  "$4\r\na\r\nb\r\n",
		},
		{
			name:  "array empty",
			write: func(w *Writer) error { return w.WriteArray(0) },
			want:  "*0\r\n",
		},
		{
			name:  "array null",
			write: func(w *Writer) error { return w.WriteArray(-1) },
			want:  "*-1\r\n",
		},
		{
			name:    "array invalid length",
			write:   func(w *Writer) error { return w.WriteArray(-2) },
			wantErr: ErrInvalidValue,
		},
		{
			name: "array two bulks",
			write: func(w *Writer) error {
				if err := w.WriteArray(2); err != nil {
					return err
				}
				if err := w.WriteBulk([]byte("PING")); err != nil {
					return err
				}
				return w.WriteBulk([]byte("x"))
			},
			want: "*2\r\n$4\r\nPING\r\n$1\r\nx\r\n",
		},
		{
			name: "nested array",
			write: func(w *Writer) error {
				if err := w.WriteArray(2); err != nil {
					return err
				}
				if err := w.WriteInt(1); err != nil {
					return err
				}
				if err := w.WriteArray(1); err != nil {
					return err
				}
				return w.WriteString("OK")
			},
			want: "*2\r\n:1\r\n*1\r\n+OK\r\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := collect(t, tc.write)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("error = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("write: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func collect(t *testing.T, write func(*Writer) error) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	w := NewWriter(&buf)
	err := write(w)
	if err != nil {
		return "", err
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	return buf.String(), nil
}
