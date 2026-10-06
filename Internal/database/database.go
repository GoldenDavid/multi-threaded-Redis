package database

import (
	"multi-threaded-Redis/Internal/aof"
	"multi-threaded-Redis/Internal/resp"
	"strings"
	"time"
)

type DataType string

const (
	TypeString DataType = "string"
	TypeList   DataType = "list"
	TypeHash   DataType = "hash"
)

type DataEntity struct {
	Type DataType
	Val  interface{}
}

type Database struct {
	data map[string]DataEntity
	ttl  map[string]int64
	aof  *aof.Aof
}

func NewDatabase() *Database {
	return &Database{
		data: make(map[string]DataEntity),
		ttl:  make(map[string]int64),
	}
}

func (db *Database) SetAof(a *aof.Aof) {
	db.aof = a
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

	result := handler(db, args)

	if db.aof != nil && isWriteCommand(commandName) {
		if result.Type != "error" {
			db.aof.Write(cmd)
		}
	}

	return result
}

func isWriteCommand(cmdName string) bool {
	writeCommands := map[string]bool{
		"SET":   true,
		"DEL":   true,
		"LPUSH": true,
		"RPUSH": true,
		"LPOP":  true,
		"RPOP":   true,
		"HSET":   true,
		"HDEL":   true,
		"EXPIRE": true,
	}
	return writeCommands[cmdName]
}

func (db *Database) IsExpired(key string) bool {
	deadline, ok := db.ttl[key]
	if !ok {
		return false
	}
	if time.Now().UnixMilli() > deadline {
		return true
	}
	return false
}

func (db *Database) DeleteExpiredKeys(limit int) {
	// Active expiration: sample a few keys with TTL, delete if expired
	// Since Go maps iterate pseudo-randomly, this works well.
	count := 0
	for key, deadline := range db.ttl {
		if count >= limit {
			break
		}
		if time.Now().UnixMilli() > deadline {
			delete(db.data, key)
			delete(db.ttl, key)
			// Ideally we also append DEL to AOF here, but keeping it simple for now.
			if db.aof != nil {
				db.aof.Write(resp.Value{
					Type: "array",
					Array: []resp.Value{
						{Type: "bulk", Bulk: "DEL"},
						{Type: "bulk", Bulk: key},
					},
				})
			}
		}
		count++
	}
}
