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
