// qdis - An efficient in-memory cache in Go
// Copyright (C) 2026 - 2027 Yixiang Qiu. All rights reserved.
// Use of this source code is governed by GPL-3. See LICENSE.

package store

import (
	"errors"
	"strconv"
	"sync"
	"testing"
)

func TestStoreGetSetDelete(t *testing.T) {
	tests := []struct {
		name    string
		run     func(*Store) (string, error)
		want    string
		wantErr error
	}{
		{
			name: "missing",
			run: func(s *Store) (string, error) {
				return s.Get("k")
			},
			wantErr: ErrKeyNotFound,
		},
		{
			name: "set then get",
			run: func(s *Store) (string, error) {
				s.Set("k", "v")
				return s.Get("k")
			},
			want: "v",
		},
		{
			name: "overwrite",
			run: func(s *Store) (string, error) {
				s.Set("k", "a")
				s.Set("k", "b")
				return s.Get("k")
			},
			want: "b",
		},
		{
			name: "delete missing then get",
			run: func(s *Store) (string, error) {
				s.Delete("k")
				return s.Get("k")
			},
			wantErr: ErrKeyNotFound,
		},
		{
			name: "delete existing",
			run: func(s *Store) (string, error) {
				s.Set("k", "v")
				s.Delete("k")
				return s.Get("k")
			},
			wantErr: ErrKeyNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.run(NewStore())
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStoreConcurrent(_ *testing.T) {
	s := NewStore()
	const goroutines = 8
	const iters = 200
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()
			key := strconv.Itoa(id % 4)
			val := strconv.Itoa(id)
			for i := 0; i < iters; i++ {
				s.Set(key, val)
				_, err := s.Get(key)
				if err != nil && !errors.Is(err, ErrKeyNotFound) {
					return
				}
				s.Delete(key)
			}
		}(g)
	}
	wg.Wait()
}
