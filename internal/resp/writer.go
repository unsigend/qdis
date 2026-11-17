// qdis - An efficient in-memory cache in Go
// Copyright (C) 2026 - 2027 Yixiang Qiu. All rights reserved.
// Use of this source code is governed by GPL-3. See LICENSE.

// Package resp implements the redis RESP (REdis Serialization Protocol),
// a simple protocol for serializing and deserializing data in Redis.
// Reference:  https://redis.io/docs/latest/develop/reference/protocol-spec/
package resp

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

// Writer is a writer that writes RESP values to a bufio.Writer.
type Writer struct {
	w   *bufio.Writer
	buf []byte
}

// NewWriter creates a new RESP writer.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:   bufio.NewWriter(w),
		buf: make([]byte, 64),
	}
}

// Flush writes any buffered data to the underlying io.Writer.
func (w *Writer) Flush() error {
	return w.w.Flush()
}

// WriteString writes a simple string to the RESP writer.
func (w *Writer) WriteString(s string) error {
	if strings.Contains(s, "\r") || strings.Contains(s, "\n") {
		return ErrInvalidValue
	}

	_, err := w.w.WriteString(string(simpleStringPrefix) + s + terminator)
	return err
}

// WriteError writes an error to the RESP writer.
func (w *Writer) WriteError(s string) error {
	if strings.Contains(s, "\r") || strings.Contains(s, "\n") {
		return ErrInvalidValue
	}

	_, err := w.w.WriteString(string(errorPrefix) + s + terminator)
	return err
}

// WriteInt writes an integer to the RESP writer.
func (w *Writer) WriteInt(i int64) error {
	str := strconv.FormatInt(i, 10)
	_, err := w.w.WriteString(string(integerPrefix) + str + terminator)
	return err
}

// WriteBulk writes a bulk string to the RESP writer.
// Pass nil to write a null bulk string.
func (w *Writer) WriteBulk(s []byte) error {
	if s == nil {
		_, err := w.w.WriteString(string(bulkStringPrefix) + "-1" + terminator)
		return err
	}

	l := strconv.Itoa(len(s))
	_, err := w.w.WriteString(string(bulkStringPrefix) + l + terminator + string(s) + terminator)
	return err
}

// WriteArray writes an array header to the RESP writer.
// Pass -1 to write a null array.
func (w *Writer) WriteArray(length int) error {
	if length == -1 {
		_, err := w.w.WriteString(string(arrayPrefix) + "-1" + terminator)
		return err
	}
	if length < -1 {
		return ErrInvalidValue
	}

	l := strconv.Itoa(length)
	_, err := w.w.WriteString(string(arrayPrefix) + l + terminator)
	return err
}
