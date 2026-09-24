package database_test

import (
	"multi-threaded-Redis/Internal/database"
	"multi-threaded-Redis/Internal/resp"
	"os"
	"testing"
)

func TestRDBSnapshotting(t *testing.T) {
	db := database.NewDatabase()

	// Populate DB
	db.Exec(resp.Value{Type: "array", Array: []resp.Value{{Type: "bulk", Bulk: "SET"}, {Type: "bulk", Bulk: "str_key"}, {Type: "bulk", Bulk: "str_val"}}})
	db.Exec(resp.Value{Type: "array", Array: []resp.Value{{Type: "bulk", Bulk: "LPUSH"}, {Type: "bulk", Bulk: "list_key"}, {Type: "bulk", Bulk: "list_val"}}})
	db.Exec(resp.Value{Type: "array", Array: []resp.Value{{Type: "bulk", Bulk: "HSET"}, {Type: "bulk", Bulk: "hash_key"}, {Type: "bulk", Bulk: "field"}, {Type: "bulk", Bulk: "hash_val"}}})

	// SAVE
	res := db.Exec(resp.Value{Type: "array", Array: []resp.Value{{Type: "bulk", Bulk: "SAVE"}}})
	if res.Type != "string" || res.Str != "OK" {
		t.Errorf("Expected OK, got %v", res)
	}

	// Verify file exists
	if _, err := os.Stat("dump.rdb"); os.IsNotExist(err) {
		t.Errorf("dump.rdb was not created")
	}

	// Create a new DB and load RDB
	newDB := database.NewDatabase()
	err := newDB.LoadRDB("dump.rdb")
	if err != nil {
		t.Errorf("Failed to load RDB: %v", err)
	}

	// Verify data
	res = newDB.Exec(resp.Value{Type: "array", Array: []resp.Value{{Type: "bulk", Bulk: "GET"}, {Type: "bulk", Bulk: "str_key"}}})
	if res.Type != "bulk" || res.Bulk != "str_val" {
		t.Errorf("Expected str_val, got %v", res)
	}

	res = newDB.Exec(resp.Value{Type: "array", Array: []resp.Value{{Type: "bulk", Bulk: "LPOP"}, {Type: "bulk", Bulk: "list_key"}}})
	if res.Type != "bulk" || res.Bulk != "list_val" {
		t.Errorf("Expected list_val, got %v", res)
	}

	res = newDB.Exec(resp.Value{Type: "array", Array: []resp.Value{{Type: "bulk", Bulk: "HGET"}, {Type: "bulk", Bulk: "hash_key"}, {Type: "bulk", Bulk: "field"}}})
	if res.Type != "bulk" || res.Bulk != "hash_val" {
		t.Errorf("Expected hash_val, got %v", res)
	}

	// Clean up
	os.Remove("dump.rdb")
}
