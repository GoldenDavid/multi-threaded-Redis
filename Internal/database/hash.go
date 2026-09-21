package database

import (
	"multi-threaded-Redis/Internal/resp"
)

func execHSet(db *Database, args []resp.Value) resp.Value {
	if len(args) < 3 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'hset' command"}
	}
	key := args[0].Bulk
	field := args[1].Bulk
	val := args[2].Bulk

	entity, exists := db.data[key]
	if !exists {
		entity = DataEntity{
			Type: TypeHash,
			Val:  make(map[string]string),
		}
		db.data[key] = entity
	}

	if entity.Type != TypeHash {
		return resp.Value{Type: "error", Str: "WRONGTYPE Operation against a key holding the wrong kind of value"}
	}

	hash := entity.Val.(map[string]string)
	
	// Check if field exists to return 0 or 1
	var result int
	if _, fieldExists := hash[field]; fieldExists {
		result = 0
	} else {
		result = 1
	}

	hash[field] = val
	return resp.Value{Type: "integer", Num: result}
}

func execHGet(db *Database, args []resp.Value) resp.Value {
	if len(args) != 2 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'hget' command"}
	}
	key := args[0].Bulk
	field := args[1].Bulk

	entity, exists := db.data[key]
	if !exists {
		return resp.Value{Type: "null"}
	}

	if entity.Type != TypeHash {
		return resp.Value{Type: "error", Str: "WRONGTYPE Operation against a key holding the wrong kind of value"}
	}

	hash := entity.Val.(map[string]string)
	
	if val, fieldExists := hash[field]; fieldExists {
		return resp.Value{Type: "bulk", Bulk: val}
	}

	return resp.Value{Type: "null"}
}

func execHGetAll(db *Database, args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'hgetall' command"}
	}
	key := args[0].Bulk

	entity, exists := db.data[key]
	if !exists {
		return resp.Value{Type: "array", Array: []resp.Value{}}
	}

	if entity.Type != TypeHash {
		return resp.Value{Type: "error", Str: "WRONGTYPE Operation against a key holding the wrong kind of value"}
	}

	hash := entity.Val.(map[string]string)
	
	res := make([]resp.Value, 0, len(hash)*2)
	for field, val := range hash {
		res = append(res, resp.Value{Type: "bulk", Bulk: field})
		res = append(res, resp.Value{Type: "bulk", Bulk: val})
	}

	return resp.Value{Type: "array", Array: res}
}

func execHDel(db *Database, args []resp.Value) resp.Value {
	if len(args) < 2 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'hdel' command"}
	}
	key := args[0].Bulk

	entity, exists := db.data[key]
	if !exists {
		return resp.Value{Type: "integer", Num: 0}
	}

	if entity.Type != TypeHash {
		return resp.Value{Type: "error", Str: "WRONGTYPE Operation against a key holding the wrong kind of value"}
	}

	hash := entity.Val.(map[string]string)
	
	deletedCount := 0
	for i := 1; i < len(args); i++ {
		field := args[i].Bulk
		if _, fieldExists := hash[field]; fieldExists {
			delete(hash, field)
			deletedCount++
		}
	}

	return resp.Value{Type: "integer", Num: deletedCount}
}
