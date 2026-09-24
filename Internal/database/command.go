package database

import (
	"multi-threaded-Redis/Internal/resp"
)

type CommandFunc func(db *Database, args []resp.Value) resp.Value

var commands = map[string]CommandFunc{
	"PING":    execPing,
	"GET":     execGet,
	"SET":     execSet,
	"DEL":     execDel,
	"LPUSH":   execLPush,
	"RPUSH":   execRPush,
	"LPOP":    execLPop,
	"RPOP":    execRPop,
	"LRANGE":  execLRange,
	"HSET":    execHSet,
	"HGET":    execHGet,
	"HGETALL": execHGetAll,
	"HDEL":    execHDel,
	"SAVE":    execSave,
}

func execSave(db *Database, args []resp.Value) resp.Value {
	if len(args) != 0 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'save' command"}
	}
	err := db.SaveRDB("dump.rdb")
	if err != nil {
		return resp.Value{Type: "error", Str: "ERR " + err.Error()}
	}
	return resp.Value{Type: "string", Str: "OK"}
}

func execPing(db *Database, args []resp.Value) resp.Value {
	if len(args) == 0 {
		return resp.Value{Type: "string", Str: "PONG"}
	}
	if len(args) == 1 {
		return resp.Value{Type: "bulk", Bulk: args[0].Bulk}
	}
	return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'ping' command"}
}

