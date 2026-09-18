package database

import (
	"multi-threaded-Redis/Internal/resp"
)

func execGet(db *Database, args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'get' command"}
	}
	key := args[0].Bulk

	val, exists := db.data[key]

	if !exists {
		return resp.Value{Type: "null"}
	}
	return val
}

func execSet(db *Database, args []resp.Value) resp.Value {
	if len(args) < 2 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'set' command"}
	}
	key := args[0].Bulk
	val := args[1]

	db.data[key] = val

	return resp.Value{Type: "string", Str: "OK"}
}

func execDel(db *Database, args []resp.Value) resp.Value {
	if len(args) < 1 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'del' command"}
	}

	deletedCount := 0
	for _, arg := range args {
		key := arg.Bulk
		if _, exists := db.data[key]; exists {
			delete(db.data, key)
			deletedCount++
		}
	}

	return resp.Value{Type: "integer", Num: deletedCount}
}
