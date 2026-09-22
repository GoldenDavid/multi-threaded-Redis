package database_test

import (
	"multi-threaded-Redis/Internal/database"
	"multi-threaded-Redis/Internal/resp"
	"testing"
)

func TestStringCommands(t *testing.T) {
	db := database.NewDatabase()
	
	// SET key val
	res := db.Exec(resp.Value{Type: "array", Array: []resp.Value{{Type: "bulk", Bulk: "SET"}, {Type: "bulk", Bulk: "mykey"}, {Type: "bulk", Bulk: "myval"}}})
	if res.Type != "string" || res.Str != "OK" {
		t.Errorf("Expected OK, got %v", res)
	}

	// GET key
	res = db.Exec(resp.Value{Type: "array", Array: []resp.Value{{Type: "bulk", Bulk: "GET"}, {Type: "bulk", Bulk: "mykey"}}})
	if res.Type != "bulk" || res.Bulk != "myval" {
		t.Errorf("Expected myval, got %v", res)
	}

	// DEL key
	res = db.Exec(resp.Value{Type: "array", Array: []resp.Value{{Type: "bulk", Bulk: "DEL"}, {Type: "bulk", Bulk: "mykey"}}})
	if res.Type != "integer" || res.Num != 1 {
		t.Errorf("Expected 1, got %v", res)
	}
}

func TestListCommands(t *testing.T) {
	db := database.NewDatabase()

	// LPUSH mylist val1
	res := db.Exec(resp.Value{Type: "array", Array: []resp.Value{{Type: "bulk", Bulk: "LPUSH"}, {Type: "bulk", Bulk: "mylist"}, {Type: "bulk", Bulk: "val1"}}})
	if res.Type != "integer" || res.Num != 1 {
		t.Errorf("Expected 1, got %v", res)
	}

	// RPUSH mylist val2
	res = db.Exec(resp.Value{Type: "array", Array: []resp.Value{{Type: "bulk", Bulk: "RPUSH"}, {Type: "bulk", Bulk: "mylist"}, {Type: "bulk", Bulk: "val2"}}})
	if res.Type != "integer" || res.Num != 2 {
		t.Errorf("Expected 2, got %v", res)
	}

	// LPOP mylist
	res = db.Exec(resp.Value{Type: "array", Array: []resp.Value{{Type: "bulk", Bulk: "LPOP"}, {Type: "bulk", Bulk: "mylist"}}})
	if res.Type != "bulk" || res.Bulk != "val1" {
		t.Errorf("Expected val1, got %v", res)
	}

	// LRANGE mylist 0 -1
	res = db.Exec(resp.Value{Type: "array", Array: []resp.Value{{Type: "bulk", Bulk: "LRANGE"}, {Type: "bulk", Bulk: "mylist"}, {Type: "bulk", Bulk: "0"}, {Type: "bulk", Bulk: "-1"}}})
	if res.Type != "array" || len(res.Array) != 1 || res.Array[0].Bulk != "val2" {
		t.Errorf("Expected [val2], got %v", res)
	}
}

func TestHashCommands(t *testing.T) {
	db := database.NewDatabase()

	// HSET myhash field1 val1
	res := db.Exec(resp.Value{Type: "array", Array: []resp.Value{{Type: "bulk", Bulk: "HSET"}, {Type: "bulk", Bulk: "myhash"}, {Type: "bulk", Bulk: "field1"}, {Type: "bulk", Bulk: "val1"}}})
	if res.Type != "integer" || res.Num != 1 {
		t.Errorf("Expected 1, got %v", res)
	}

	// HGET myhash field1
	res = db.Exec(resp.Value{Type: "array", Array: []resp.Value{{Type: "bulk", Bulk: "HGET"}, {Type: "bulk", Bulk: "myhash"}, {Type: "bulk", Bulk: "field1"}}})
	if res.Type != "bulk" || res.Bulk != "val1" {
		t.Errorf("Expected val1, got %v", res)
	}

	// HGETALL myhash
	res = db.Exec(resp.Value{Type: "array", Array: []resp.Value{{Type: "bulk", Bulk: "HGETALL"}, {Type: "bulk", Bulk: "myhash"}}})
	if res.Type != "array" || len(res.Array) != 2 {
		t.Errorf("Expected array of len 2, got %v", res)
	}

	// HDEL myhash field1
	res = db.Exec(resp.Value{Type: "array", Array: []resp.Value{{Type: "bulk", Bulk: "HDEL"}, {Type: "bulk", Bulk: "myhash"}, {Type: "bulk", Bulk: "field1"}}})
	if res.Type != "integer" || res.Num != 1 {
		t.Errorf("Expected 1, got %v", res)
	}
}
