// qdis - An efficient in-memory cache in Go
// Copyright (C) 2026 - 2027 Yixiang Qiu. All rights reserved.
// Use of this source code is governed by GPL-3. See LICENSE.

package server

import (
	"errors"
	"strconv"
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
func (c *Command) set(w *resp.Writer, st *store.Store) error {
	if len(c.Args) != 2 {
		return w.WriteError("ERR wrong number of arguments for 'set' command")
	}
	st.Set(c.Args[0], c.Args[1])
	return w.WriteString("OK")
}

// get runs the GET command.
func (c *Command) get(w *resp.Writer, st *store.Store) error {
	if len(c.Args) != 1 {
		return w.WriteError("ERR wrong number of arguments for 'get' command")
	}
	v, err := st.Get(c.Args[0])
	if err != nil {
		return w.WriteBulk(nil)
	}
	return w.WriteBulk([]byte(v))
}

// del runs the DEL command.
func (c *Command) del(w *resp.Writer, st *store.Store) error {
	if len(c.Args) < 1 {
		return w.WriteError("ERR wrong number of arguments for 'del' command")
	}
	n := st.DeleteMany(c.Args...)
	return w.WriteInt(int64(n))
}

// exists runs the EXISTS command.
func (c *Command) exists(w *resp.Writer, st *store.Store) error {
	if len(c.Args) < 1 {
		return w.WriteError("ERR wrong number of arguments for 'exists' command")
	}
	n := st.Exists(c.Args...)
	return w.WriteInt(int64(n))
}

// typeCmd runs the TYPE command.
func (c *Command) typeCmd(w *resp.Writer, st *store.Store) error {
	if len(c.Args) != 1 {
		return w.WriteError("ERR wrong number of arguments for 'type' command")
	}
	_, err := st.Get(c.Args[0])
	if err != nil {
		return w.WriteString("none")
	}
	return w.WriteString("string")
}

// dbsize runs the DBSIZE command.
func (c *Command) dbsize(w *resp.Writer, st *store.Store) error {
	if len(c.Args) != 0 {
		return w.WriteError("ERR wrong number of arguments for 'dbsize' command")
	}
	return w.WriteInt(int64(st.Size()))
}

// flushdb runs the FLUSHDB command.
func (c *Command) flushdb(w *resp.Writer, st *store.Store) error {
	if len(c.Args) != 0 {
		return w.WriteError("ERR syntax error")
	}
	st.Drop()
	return w.WriteString("OK")
}

// keys runs the KEYS command.
func (c *Command) keys(w *resp.Writer, st *store.Store) error {
	if len(c.Args) != 1 {
		return w.WriteError("ERR wrong number of arguments for 'keys' command")
	}
	pattern := c.Args[0]
	keys, err := st.Keys(pattern)
	if err != nil {
		return w.WriteError("ERR invalid pattern")
	}
	err = w.WriteArray(len(keys))
	if err != nil {
		return err
	}
	for _, key := range keys {
		err = w.WriteBulk([]byte(key))
		if err != nil {
			return err
		}
	}
	return nil
}

// incr runs the INCR command.
func (c *Command) incr(w *resp.Writer, st *store.Store) error {
	if len(c.Args) != 1 {
		return w.WriteError("ERR wrong number of arguments for 'incr' command")
	}
	n, err := st.Incr(c.Args[0], 1)
	if err != nil {
		if errors.Is(err, store.ErrNotInt) {
			return w.WriteError("ERR value is not an integer or out of range")
		}
		if errors.Is(err, store.ErrOverflow) {
			return w.WriteError("ERR increment or decrement would overflow")
		}
		return err
	}
	return w.WriteInt(n)
}

// incrby runs the INCRBY command.
func (c *Command) incrby(w *resp.Writer, st *store.Store) error {
	if len(c.Args) != 2 {
		return w.WriteError("ERR wrong number of arguments for 'incrby' command")
	}

	// check if the value is a valid integer
	// +8, 08, -08, 008, etc. are invalid
	if strings.HasPrefix(c.Args[1], "+") ||
		(len(c.Args[1]) > 1 && c.Args[1][0] == '0') ||
		(len(c.Args[1]) > 1 && c.Args[1][0] == '-' && c.Args[1][1] == '0') {
		return w.WriteError("ERR value is not an integer or out of range")
	}

	delta, err := strconv.ParseInt(c.Args[1], 10, 64)
	if err != nil {
		return w.WriteError("ERR value is not an integer or out of range")
	}
	n, err := st.Incr(c.Args[0], delta)
	if err != nil {
		if errors.Is(err, store.ErrNotInt) {
			return w.WriteError("ERR value is not an integer or out of range")
		}
		if errors.Is(err, store.ErrOverflow) {
			return w.WriteError("ERR increment or decrement would overflow")
		}
		return err
	}
	return w.WriteInt(n)
}

// append runs the APPEND command.
func (c *Command) append(w *resp.Writer, st *store.Store) error {
	if len(c.Args) != 2 {
		return w.WriteError("ERR wrong number of arguments for 'append' command")
	}
	n, err := st.Append(c.Args[0], c.Args[1])
	if err != nil {
		return err
	}
	return w.WriteInt(int64(n))
}

// mset runs the MSET command.
func (c *Command) mset(w *resp.Writer, st *store.Store) error {
	if len(c.Args) == 0 || len(c.Args)%2 != 0 {
		return w.WriteError("ERR wrong number of arguments for 'mset' command")
	}

	keys := make([]string, 0, len(c.Args)/2)
	values := make([]string, 0, len(c.Args)/2)

	for i := 0; i < len(c.Args); i += 2 {
		keys = append(keys, c.Args[i])
		values = append(values, c.Args[i+1])
	}

	// must succeed
	err := st.SetMany(keys, values)
	if err != nil {
		return err
	}

	return w.WriteString("OK")
}

// mget runs the MGET command.
func (c *Command) mget(w *resp.Writer, st *store.Store) error {
	if len(c.Args) == 0 {
		return w.WriteError("ERR wrong number of arguments for 'mget' command")
	}
	values, exists := st.GetMany(c.Args...)

	err := w.WriteArray(len(values))
	if err != nil {
		return err
	}

	for i, value := range values {
		if exists[i] {
			err = w.WriteBulk([]byte(value))
			if err != nil {
				return err
			}
		} else {
			err = w.WriteBulk(nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// cmdtable is a table of commands.
var cmdtable = map[string]func(*Command, *resp.Writer, *store.Store) error{
	"PING":    (*Command).ping,
	"ECHO":    (*Command).echo,
	"QUIT":    (*Command).quit,
	"SET":     (*Command).set,
	"GET":     (*Command).get,
	"DEL":     (*Command).del,
	"EXISTS":  (*Command).exists,
	"TYPE":    (*Command).typeCmd,
	"DBSIZE":  (*Command).dbsize,
	"FLUSHDB": (*Command).flushdb,
	"KEYS":    (*Command).keys,
	"INCR":    (*Command).incr,
	"INCRBY":  (*Command).incrby,
	"APPEND":  (*Command).append,
	"MSET":    (*Command).mset,
	"MGET":    (*Command).mget,
}

// RunCommand runs a command.
func (c *Command) RunCommand(w *resp.Writer, store *store.Store) error {
	cmd, ok := cmdtable[strings.ToUpper(c.Name)]
	if !ok {
		return ErrInvalidCommand
	}
	return cmd(c, w, store)
}
