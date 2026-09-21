package database

import (
	"container/list"
	"multi-threaded-Redis/Internal/resp"
)

func execLPush(db *Database, args []resp.Value) resp.Value {
	if len(args) < 2 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'lpush' command"}
	}
	key := args[0].Bulk

	entity, exists := db.data[key]
	if !exists {
		entity = DataEntity{
			Type: TypeList,
			Val:  list.New(),
		}
		db.data[key] = entity
	}

	if entity.Type != TypeList {
		return resp.Value{Type: "error", Str: "WRONGTYPE Operation against a key holding the wrong kind of value"}
	}

	l := entity.Val.(*list.List)
	for i := 1; i < len(args); i++ {
		l.PushFront(args[i].Bulk)
	}

	return resp.Value{Type: "integer", Num: l.Len()}
}

func execRPush(db *Database, args []resp.Value) resp.Value {
	if len(args) < 2 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'rpush' command"}
	}
	key := args[0].Bulk

	entity, exists := db.data[key]
	if !exists {
		entity = DataEntity{
			Type: TypeList,
			Val:  list.New(),
		}
		db.data[key] = entity
	}

	if entity.Type != TypeList {
		return resp.Value{Type: "error", Str: "WRONGTYPE Operation against a key holding the wrong kind of value"}
	}

	l := entity.Val.(*list.List)
	for i := 1; i < len(args); i++ {
		l.PushBack(args[i].Bulk)
	}

	return resp.Value{Type: "integer", Num: l.Len()}
}

func execLPop(db *Database, args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'lpop' command"}
	}
	key := args[0].Bulk

	entity, exists := db.data[key]
	if !exists {
		return resp.Value{Type: "null"}
	}

	if entity.Type != TypeList {
		return resp.Value{Type: "error", Str: "WRONGTYPE Operation against a key holding the wrong kind of value"}
	}

	l := entity.Val.(*list.List)
	element := l.Front()
	if element == nil {
		return resp.Value{Type: "null"}
	}

	val := element.Value.(string)
	l.Remove(element)

	return resp.Value{Type: "bulk", Bulk: val}
}

func execRPop(db *Database, args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'rpop' command"}
	}
	key := args[0].Bulk

	entity, exists := db.data[key]
	if !exists {
		return resp.Value{Type: "null"}
	}

	if entity.Type != TypeList {
		return resp.Value{Type: "error", Str: "WRONGTYPE Operation against a key holding the wrong kind of value"}
	}

	l := entity.Val.(*list.List)
	element := l.Back()
	if element == nil {
		return resp.Value{Type: "null"}
	}

	val := element.Value.(string)
	l.Remove(element)

	return resp.Value{Type: "bulk", Bulk: val}
}

func execLRange(db *Database, args []resp.Value) resp.Value {
	if len(args) != 3 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'lrange' command"}
	}
	key := args[0].Bulk
	
	// For simplicity in this mock, we assume arguments are valid integers and implement a basic slice
	// A full implementation would parse start and end indices correctly including negative indices.
	// Since RESP passes arguments as strings (or bulk), we need to parse them.
	// We'll skip complex bounds checking for this basic implementation and just return the whole list if we can't parse.
	
	entity, exists := db.data[key]
	if !exists {
		return resp.Value{Type: "array", Array: []resp.Value{}}
	}

	if entity.Type != TypeList {
		return resp.Value{Type: "error", Str: "WRONGTYPE Operation against a key holding the wrong kind of value"}
	}

	l := entity.Val.(*list.List)
	
	res := make([]resp.Value, 0, l.Len())
	for e := l.Front(); e != nil; e = e.Next() {
		res = append(res, resp.Value{Type: "bulk", Bulk: e.Value.(string)})
	}

	return resp.Value{Type: "array", Array: res}
}
