// qdis - An efficient in-memory cache in Go
// Copyright (C) 2026 - 2027 Yixiang Qiu. All rights reserved.
// Use of this source code is governed by GPL-3. See LICENSE.

package server

import (
	"errors"
	"strings"

	"github.com/unsigend/qdis/internal/resp"
	"github.com/unsigend/qdis/internal/store"
)

// ErrInvalidCommand is returned when a command is invalid.
var ErrInvalidCommand = errors.New("invalid command")

// ErrInvalidValue is returned when a value is invalid.
var ErrInvalidValue = errors.New("invalid value")

// ErrQuit is returned when the QUIT command is received.
var ErrQuit = errors.New("quit")

// Command is a command.
type Command struct {
	Name string
	Args []string
}

// NewCommand creates a new command.
func NewCommand(name string, args ...string) *Command {
	return &Command{
		Name: name,
		Args: args,
	}
}

// NewCommandRaw creates a new command from a raw command.
// Assume the first element is the command name, and the rest are the arguments.
func NewCommandRaw(raw [][]byte) (*Command, error) {
	if len(raw) == 0 {
		return nil, ErrInvalidValue
	}
	var cmd Command
	cmd.Name = string(raw[0])
	cmd.Args = make([]string, len(raw)-1)
	for i := 1; i < len(raw); i++ {
		cmd.Args[i-1] = string(raw[i])
	}
	return &cmd, nil
}

// ping runs the PING command.
func (c *Command) ping(w *resp.Writer, _ *store.Store) error {
	if len(c.Args) == 0 {
		return w.WriteString("PONG")
	}
	if len(c.Args) > 1 {
		return w.WriteError("ERR wrong number of arguments for 'ping' command")
	}
	return w.WriteBulk([]byte(c.Args[0]))
}

// echo runs the ECHO command.
func (c *Command) echo(w *resp.Writer, _ *store.Store) error {
	if len(c.Args) != 1 {
		return w.WriteError("ERR wrong number of arguments for 'echo' command")
	}
	return w.WriteBulk([]byte(c.Args[0]))
}

// quit runs the QUIT command.
func (c *Command) quit(w *resp.Writer, _ *store.Store) error {
	if err := w.WriteString("OK"); err != nil {
		return err
	}
	return ErrQuit
}

// set runs the SET command.
func (c *Command) set(w *resp.Writer, store *store.Store) error {
	if len(c.Args) != 2 {
		return w.WriteError("ERR wrong number of arguments for 'set' command")
	}
	store.Set(c.Args[0], c.Args[1])
	return w.WriteString("OK")
}

// get runs the GET command.
func (c *Command) get(w *resp.Writer, store *store.Store) error {
	if len(c.Args) != 1 {
		return w.WriteError("ERR wrong number of arguments for 'get' command")
	}
	v, err := store.Get(c.Args[0])
	if err != nil {
		return w.WriteBulk(nil)
	}
	return w.WriteBulk([]byte(v))
}

// cmdtable is a table of commands.
var cmdtable = map[string]func(*Command, *resp.Writer, *store.Store) error{
	"PING": (*Command).ping,
	"ECHO": (*Command).echo,
	"QUIT": (*Command).quit,
	"SET":  (*Command).set,
	"GET":  (*Command).get,
}

// RunCommand runs a command.
func (c *Command) RunCommand(w *resp.Writer, store *store.Store) error {
	cmd, ok := cmdtable[strings.ToUpper(c.Name)]
	if !ok {
		return ErrInvalidCommand
	}
	return cmd(c, w, store)
}
