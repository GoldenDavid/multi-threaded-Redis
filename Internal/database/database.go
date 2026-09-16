package database

import (
	"multi-threaded-Redis/Internal/resp"
	"strings"
	"sync"
)

type Database struct {
	data map[string]resp.Value
	mu   sync.RWMutex
}

func NewDatabase() *Database {
	return &Database{
		data: make(map[string]resp.Value),
	}
}

func (db *Database) Exec(cmd resp.Value) resp.Value {
	if cmd.Type != "array" || len(cmd.Array) == 0 {
		return resp.Value{Type: "error", Str: "ERR invalid request"}
	}

	commandName := strings.ToUpper(cmd.Array[0].Bulk)
	args := cmd.Array[1:]

	handler, ok := commands[commandName]
	if !ok {
		return resp.Value{Type: "error", Str: "ERR unknown command '" + commandName + "'"}
	}

	return handler(db, args)
}
