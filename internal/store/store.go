// qdis - An efficient in-memory cache in Go
// Copyright (C) 2026 - 2027 Yixiang Qiu. All rights reserved.
// Use of this source code is governed by GPL-3. See LICENSE.

// Package store implements the store for the qdis server.
package store

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// ErrKeyNotFound is returned when a key is not found.
var ErrKeyNotFound = errors.New("key not found")

// ErrInvalidPattern is returned when a pattern is invalid.
var ErrInvalidPattern = errors.New("invalid pattern")

// ErrNotInt is returned when a value is not an integer.
var ErrNotInt = errors.New("not an integer")

// ErrOverflow is returned when an integer overflows.
var ErrOverflow = errors.New("integer overflow")

// ErrUnbalancedKV is returned when the number of keys and values is not equal.
var ErrUnbalancedKV = errors.New("unbalanced key-value pairs")

// Store represents the key-value store.
type Store struct {
	db map[string]string
	mu sync.Mutex // removed when executer is implemented
}

// NewStore creates a new store.
func NewStore() *Store {
	return &Store{
		db: make(map[string]string),
		mu: sync.Mutex{},
	}
}

// Set sets a value in the store.
func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.db[key] = value
}

// SetMany sets multiple values in the store.
// Returns: An error if the number of keys and values is not equal.
func (s *Store) SetMany(keys []string, values []string) error {
	if len(keys) != len(values) {
		return ErrUnbalancedKV
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range keys {
		s.db[keys[i]] = values[i]
	}
	return nil
}

// Get gets a value from the store.
func (s *Store) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.db[key]
	if !ok {
		return "", ErrKeyNotFound
	}
	return value, nil
}

// GetMany gets multiple values from the store.
// Returns: A slice of values and a slice of booleans indicating if the key exists.
func (s *Store) GetMany(keys ...string) ([]string, []bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	values := make([]string, 0, len(keys))
	exists := make([]bool, 0, len(keys))

	for _, key := range keys {
		value, ok := s.db[key]
		if !ok {
			values = append(values, "")
			exists = append(exists, false)
		} else {
			values = append(values, value)
			exists = append(exists, true)
		}
	}
	return values, exists
}

// Delete deletes a key from the store.
func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.db, key)
}

// DeleteMany deletes multiple keys from the store.
// It returns the number of keys deleted.
func (s *Store) DeleteMany(keys ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, key := range keys {
		if _, ok := s.db[key]; ok {
			delete(s.db, key)
			n++
		}
	}
	return n
}

// Exists checks if multiple keys exist in the store.
// It returns the number of keys that exist.
func (s *Store) Exists(keys ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, key := range keys {
		if _, ok := s.db[key]; ok {
			n++
		}
	}
	return n
}

// Size returns the number of keys in the store.
func (s *Store) Size() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.db)
}

// Drop drops the store.
func (s *Store) Drop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.db = make(map[string]string)
}

// Keys returns all keys that match the pattern.
// Pattern: *, ?, [x]
// Returns: Array of keys that match the pattern.
// Error: ErrInvalidPattern if the pattern is invalid.
func (s *Store) Keys(pattern string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// normalize the pattern to a regex pattern
	out := "^"
	i := 0
	for i < len(pattern) {
		switch pattern[i] {
		case '*':
			out += ".*"
		case '?':
			out += "."
		case '[':
			// find the closing bracket
			end := strings.IndexByte(pattern[i:], ']')
			if end == -1 {
				return nil, ErrInvalidPattern
			}
			out += "[" + pattern[i+1:i+end] + "]"
			i += end
		default:
			out += regexp.QuoteMeta(pattern[i : i+1])
		}
		i++
	}
	out += "$"

	re, err := regexp.Compile(out)
	if err != nil {
		return nil, ErrInvalidPattern
	}

	keys := make([]string, 0, len(s.db))

	for key := range s.db {
		if re.MatchString(key) {
			keys = append(keys, key)
		}
	}
	return keys, nil

}

// Incr increments the value of a key by delta.
// If the key does not exist, it is set to 0.
// Returns: The new value of the key.
// Error: ErrNotInt if the value is not an integer.
func (s *Store) Incr(key string, delta int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.db[key]
	if !ok {
		v = "0"
	}

	// check if the value is a valid integer
	// +8, 08, -08, 008, etc. are invalid
	if strings.HasPrefix(v, "+") ||
		(len(v) > 1 && v[0] == '0') ||
		(len(v) > 1 && v[0] == '-' && v[1] == '0') {
		return 0, ErrNotInt
	}

	n, err := strconv.ParseInt(v, 10, 64)

	if err != nil {
		return 0, ErrNotInt
	}

	if delta > 0 && n > math.MaxInt64-delta {
		return 0, ErrOverflow
	}

	if delta < 0 && n < math.MinInt64-delta {
		return 0, ErrOverflow
	}

	n += delta
	s.db[key] = strconv.FormatInt(n, 10)
	return n, nil
}

// Append appends a value to the end of the value of a key.
// If the key does not exist, it is set to the value.
// Returns: The new length of the value.
func (s *Store) Append(key, value string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.db[key]
	if !ok {
		v = ""
	}
	s.db[key] = v + value
	return len(s.db[key]), nil
}
