package database

import (
	"multi-threaded-Redis/Internal/aof"
	"multi-threaded-Redis/Internal/resp"
	"strings"
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
	aof  *aof.Aof
}

func NewDatabase() *Database {
	return &Database{
		data: make(map[string]DataEntity),
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
		"RPOP":  true,
		"HSET":  true,
		"HDEL":  true,
	}
	return writeCommands[cmdName]
}
