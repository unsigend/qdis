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
)

// Reader is a reader that reads RESP values from an io.Reader.
type Reader struct {
	r *bufio.Reader
}

// NewReader creates a new RESP reader.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		r: bufio.NewReader(r),
	}
}

// readSegment reads a segment of a RESP value that ends with '\r\n'.
// It returns the segment data without the '\r\n'.
func (r *Reader) readSegment() ([]byte, error) {
	line, err := r.r.ReadSlice('\n')

	if err != nil {
		return nil, err
	}

	if len(line) < 2 || line[len(line)-2] != '\r' {
		return nil, ErrBadFmt
	}

	return append([]byte(nil), line[:len(line)-2]...), nil
}

// readHeader reads the header of a RESP value.
// It returns the length of the value and an error if any.
func (r *Reader) readHeader() (int, error) {
	seg, err := r.readSegment()
	if err != nil {
		return 0, err
	}

	if len(seg) < 1 {
		return 0, ErrBadFmt
	}

	if seg[0] != arrayPrefix {
		return 0, ErrInvalidValue
	}

	length, err := strconv.Atoi(string(seg[1:]))
	if err != nil {
		return 0, err
	}

	if length < 1 { // for read command it must be at least 1
		return 0, ErrInvalidValue
	}

	return length, nil
}

// ReadBulk reads a bulk string from the RESP reader.
// It returns the bulk string and an error if any.
func (r *Reader) ReadBulk() ([]byte, error) {
	seg, err := r.readSegment()
	if err != nil {
		return nil, err
	}

	if len(seg) < 1 {
		return nil, ErrBadFmt
	}

	if seg[0] != bulkStringPrefix {
		return nil, ErrInvalidValue
	}

	length, err := strconv.Atoi(string(seg[1:]))
	if err != nil {
		return nil, err
	}

	if length == -1 { // nil bulk string
		return nil, nil
	}
	if length < 0 {
		return nil, ErrInvalidValue
	}

	if length > maxBulk {
		return nil, ErrTooLargeBulk
	}

	buf := make([]byte, length)
	if _, err := io.ReadFull(r.r, buf); err != nil {
		return nil, err
	}

	crlf := make([]byte, 2)
	if _, err := io.ReadFull(r.r, crlf); err != nil {
		return nil, err
	}
	if crlf[0] != '\r' || crlf[1] != '\n' {
		return nil, ErrBadFmt
	}

	return buf, nil
}

// ReadCmd reads a command from the RESP reader.
func (r *Reader) ReadCmd() ([][]byte, error) {
	var cmd [][]byte

	length, err := r.readHeader()
	if err != nil {
		return nil, err
	}

	for range length {
		bulk, err := r.ReadBulk()
		if err != nil {
			return nil, err
		}
		cmd = append(cmd, bulk)
	}

	return cmd, nil
}

// ReadString reads a simple string from the RESP reader.
// It returns the simple string and an error if any.
func (r *Reader) ReadString() (string, error) {
	seg, err := r.readSegment()
	if err != nil {
		return "", err
	}

	if len(seg) < 1 {
		return "", ErrBadFmt
	}

	if seg[0] != simpleStringPrefix {
		return "", ErrInvalidValue
	}

	return string(seg[1:]), nil
}

// ReadError reads an error from the RESP reader.
// It returns the error message and an error if any.
func (r *Reader) ReadError() (string, error) {
	seg, err := r.readSegment()
	if err != nil {
		return "", err
	}

	if len(seg) < 1 {
		return "", ErrBadFmt
	}

	if seg[0] != errorPrefix {
		return "", ErrInvalidValue
	}

	return string(seg[1:]), nil
}

// ReadInt reads an integer from the RESP reader.
// It returns the integer and an error if any.
func (r *Reader) ReadInt() (int64, error) {
	seg, err := r.readSegment()
	if err != nil {
		return 0, err
	}

	if len(seg) < 1 {
		return 0, ErrBadFmt
	}

	if seg[0] != integerPrefix {
		return 0, ErrInvalidValue
	}

	return strconv.ParseInt(string(seg[1:]), 10, 64)
}
