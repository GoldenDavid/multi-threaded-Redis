package database

import (
	"multi-threaded-Redis/Internal/resp"
	"strconv"
	"strings"
	"time"
)

func execGet(db *Database, args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'get' command"}
	}
	key := args[0].Bulk

	if db.IsExpired(key) {
		delete(db.data, key)
		delete(db.ttl, key)
		if db.aof != nil {
			db.aof.Write(resp.Value{
				Type: "array",
				Array: []resp.Value{
					{Type: "bulk", Bulk: "DEL"},
					{Type: "bulk", Bulk: key},
				},
			})
		}
		return resp.Value{Type: "null"}
	}

	entity, exists := db.data[key]

	if !exists {
		return resp.Value{Type: "null"}
	}

	if entity.Type != TypeString {
		return resp.Value{Type: "error", Str: "WRONGTYPE Operation against a key holding the wrong kind of value"}
	}

	return resp.Value{Type: "bulk", Bulk: entity.Val.(string)}
}

func execSet(db *Database, args []resp.Value) resp.Value {
	if len(args) < 2 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'set' command"}
	}
	key := args[0].Bulk
	val := args[1].Bulk

	db.data[key] = DataEntity{
		Type: TypeString,
		Val:  val,
	}
	// Clear any existing TTL
	delete(db.ttl, key)

	// Simple EX support (e.g. SET key val EX 10)
	if len(args) == 4 {
		modifier := strings.ToUpper(args[2].Bulk)
		if modifier == "EX" {
			sec, err := strconv.ParseInt(args[3].Bulk, 10, 64)
			if err == nil {
				db.ttl[key] = time.Now().UnixMilli() + sec*1000
			}
		} else if modifier == "PX" {
			ms, err := strconv.ParseInt(args[3].Bulk, 10, 64)
			if err == nil {
				db.ttl[key] = time.Now().UnixMilli() + ms
			}
		}
	}

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
			delete(db.ttl, key)
			deletedCount++
		}
	}

	return resp.Value{Type: "integer", Num: deletedCount}
}

func execExpire(db *Database, args []resp.Value) resp.Value {
	if len(args) != 2 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'expire' command"}
	}
	key := args[0].Bulk
	
	if db.IsExpired(key) {
		delete(db.data, key)
		delete(db.ttl, key)
	}

	if _, exists := db.data[key]; !exists {
		return resp.Value{Type: "integer", Num: 0}
	}

	sec, err := strconv.ParseInt(args[1].Bulk, 10, 64)
	if err != nil {
		return resp.Value{Type: "error", Str: "ERR value is not an integer or out of range"}
	}

	db.ttl[key] = time.Now().UnixMilli() + sec*1000
	return resp.Value{Type: "integer", Num: 1}
}

func execTTL(db *Database, args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'ttl' command"}
	}
	key := args[0].Bulk

	if db.IsExpired(key) {
		delete(db.data, key)
		delete(db.ttl, key)
	}

	if _, exists := db.data[key]; !exists {
		return resp.Value{Type: "integer", Num: -2}
	}

	deadline, hasTtl := db.ttl[key]
	if !hasTtl {
		return resp.Value{Type: "integer", Num: -1}
	}

	remainMs := deadline - time.Now().UnixMilli()
	if remainMs < 0 {
		return resp.Value{Type: "integer", Num: -2}
	}

	return resp.Value{Type: "integer", Num: int(remainMs / 1000)}
}
