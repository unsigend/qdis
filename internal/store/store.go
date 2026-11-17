// qdis - An efficient in-memory cache in Go
// Copyright (C) 2026 - 2027 Yixiang Qiu. All rights reserved.
// Use of this source code is governed by GPL-3. See LICENSE.

// Package store implements the store for the qdis server.
package store

import (
	"errors"
	"sync"
)

// ErrKeyNotFound is returned when a key is not found.
var ErrKeyNotFound = errors.New("key not found")

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

// Delete deletes a key from the store.
func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.db, key)
}
