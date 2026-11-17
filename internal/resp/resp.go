// qdis - An efficient in-memory cache in Go
// Copyright (C) 2026 - 2027 Yixiang Qiu. All rights reserved.
// Use of this source code is governed by GPL-3. See LICENSE.

// Package resp implements the redis RESP (REdis Serialization Protocol),
// a simple protocol for serializing and deserializing data in Redis.
// Reference:  https://redis.io/docs/latest/develop/reference/protocol-spec/
package resp

import (
	"errors"
)

const (
	simpleStringPrefix byte = '+'
	errorPrefix        byte = '-'
	integerPrefix      byte = ':'
	bulkStringPrefix   byte = '$'
	arrayPrefix        byte = '*'
)

const terminator = "\r\n"

// ErrInvalidValue is returned when a value is invalid.
var ErrInvalidValue = errors.New("invalid value")

// ErrBadFmt is returned when a value is in bad format.
var ErrBadFmt = errors.New("bad format")

// ErrTooLargeBulk is returned when a bulk is too large.
var ErrTooLargeBulk = errors.New("bulk size too large")

const maxBulk = 512 * 1024 * 1024 // 512M
