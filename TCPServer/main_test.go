package main

import (
	"net"
	"testing"
	"time"

	"multi-threaded-Redis/Internal/resp"
)

func TestTCPServerCommands(t *testing.T) {
	// Start the server in a goroutine
	go main()

	// Give the server a moment to start up
	time.Sleep(100 * time.Millisecond)

	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	writer := resp.NewWriter(conn)
	parser := resp.NewParser(conn)

	// Helper function to send and read
	sendCommand := func(args ...string) resp.Value {
		cmdArgs := make([]resp.Value, len(args))
		for i, arg := range args {
			cmdArgs[i] = resp.Value{Type: "bulk", Bulk: arg}
		}
		
		cmd := resp.Value{
			Type:  "array",
			Array: cmdArgs,
		}

		err = writer.Write(cmd)
		if err != nil {
			t.Fatalf("Failed to write to server: %v", err)
		}

		response, err := parser.Parse()
		if err != nil {
			t.Fatalf("Failed to read from server: %v", err)
		}
		return response
	}

	// 1. Test PING
	res := sendCommand("PING")
	if res.Type != "string" || res.Str != "PONG" {
		t.Errorf("Expected PONG, got %v", res)
	}

	// 2. Test GET non-existent key
	res = sendCommand("GET", "mykey")
	if res.Type != "null" {
		t.Errorf("Expected null for GET non-existent key, got %v", res)
	}

	// 3. Test SET
	res = sendCommand("SET", "mykey", "myvalue")
	if res.Type != "string" || res.Str != "OK" {
		t.Errorf("Expected OK, got %v", res)
	}

	// 4. Test GET existing key
	res = sendCommand("GET", "mykey")
	if res.Type != "bulk" || res.Bulk != "myvalue" {
		t.Errorf("Expected bulk string myvalue, got %v", res)
	}
	
	// 5. Test DEL
	res = sendCommand("DEL", "mykey")
	if res.Type != "integer" || res.Num != 1 {
		t.Errorf("Expected integer 1, got %v", res)
	}
}

func TestConcurrentConnections(t *testing.T) {
	// The server is already started by the previous test, but if we run tests individually we might need it.
	// Since go main() blocks, in tests it's better to isolate them. For simplicity, we assume server is running on :3000.
	
	// Give the server a moment to start up if we're running everything together
	time.Sleep(100 * time.Millisecond)

	numClients := 50
	done := make(chan bool)

	for i := 0; i < numClients; i++ {
		go func(clientID int) {
			conn, err := net.Dial("tcp", "localhost:3000")
			if err != nil {
				// Retry once if server is busy
				time.Sleep(50 * time.Millisecond)
				conn, err = net.Dial("tcp", "localhost:3000")
				if err != nil {
					t.Errorf("Client %d failed to connect: %v", clientID, err)
					done <- false
					return
				}
			}
			defer conn.Close()

			writer := resp.NewWriter(conn)
			parser := resp.NewParser(conn)

			// SET a unique key
			cmd := resp.Value{
				Type: "array",
				Array: []resp.Value{
					{Type: "bulk", Bulk: "SET"},
					{Type: "bulk", Bulk: "key_" + string(rune(clientID))},
					{Type: "bulk", Bulk: "val"},
				},
			}
			
			if err := writer.Write(cmd); err != nil {
				t.Errorf("Client %d failed to write SET: %v", clientID, err)
			}
			
			if _, err := parser.Parse(); err != nil {
				t.Errorf("Client %d failed to read SET response: %v", clientID, err)
			}

			// PING
			pingCmd := resp.Value{
				Type: "array",
				Array: []resp.Value{
					{Type: "bulk", Bulk: "PING"},
				},
			}
			if err := writer.Write(pingCmd); err != nil {
				t.Errorf("Client %d failed to write PING: %v", clientID, err)
			}
			if _, err := parser.Parse(); err != nil {
				t.Errorf("Client %d failed to read PING response: %v", clientID, err)
			}

			done <- true
		}(i)
	}

	for i := 0; i < numClients; i++ {
		success := <-done
		if !success {
			t.Fatalf("A client failed")
		}
	}
}

func TestDataStructures(t *testing.T) {
	// Re-using the server started in TestTCPServerCommands is fine if tests run sequentially, 
	// but to be safe we'll use a new connection.
	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	writer := resp.NewWriter(conn)
	parser := resp.NewParser(conn)

	sendCommand := func(args ...string) resp.Value {
		cmdArgs := make([]resp.Value, len(args))
		for i, arg := range args {
			cmdArgs[i] = resp.Value{Type: "bulk", Bulk: arg}
		}
		
		cmd := resp.Value{
			Type:  "array",
			Array: cmdArgs,
		}

		err = writer.Write(cmd)
		if err != nil {
			t.Fatalf("Failed to write to server: %v", err)
		}

		response, err := parser.Parse()
		if err != nil {
			t.Fatalf("Failed to read from server: %v", err)
		}
		return response
	}

	// Clean up keys first just in case
	sendCommand("DEL", "mylist", "myhash")

	// --- LIST TESTS ---
	res := sendCommand("LPUSH", "mylist", "world")
	if res.Type != "integer" || res.Num != 1 {
		t.Errorf("Expected LPUSH to return 1, got %v", res)
	}
	
	res = sendCommand("LPUSH", "mylist", "hello")
	if res.Type != "integer" || res.Num != 2 {
		t.Errorf("Expected LPUSH to return 2, got %v", res)
	}

	res = sendCommand("RPUSH", "mylist", "!")
	if res.Type != "integer" || res.Num != 3 {
		t.Errorf("Expected RPUSH to return 3, got %v", res)
	}

	res = sendCommand("LRANGE", "mylist", "0", "-1") // Note: our lrange mock ignores indices for now
	if res.Type != "array" || len(res.Array) != 3 {
		t.Errorf("Expected array of length 3, got %v", res)
	} else {
		if res.Array[0].Bulk != "hello" || res.Array[1].Bulk != "world" || res.Array[2].Bulk != "!" {
			t.Errorf("LRANGE elements mismatch, got %v", res)
		}
	}

	res = sendCommand("LPOP", "mylist")
	if res.Type != "bulk" || res.Bulk != "hello" {
		t.Errorf("Expected LPOP to return 'hello', got %v", res)
	}

	res = sendCommand("RPOP", "mylist")
	if res.Type != "bulk" || res.Bulk != "!" {
		t.Errorf("Expected RPOP to return '!', got %v", res)
	}


	// --- HASH TESTS ---
	res = sendCommand("HSET", "myhash", "field1", "value1")
	if res.Type != "integer" || res.Num != 1 {
		t.Errorf("Expected HSET to return 1, got %v", res)
	}

	res = sendCommand("HSET", "myhash", "field1", "value2") // overwrite
	if res.Type != "integer" || res.Num != 0 {
		t.Errorf("Expected HSET overwrite to return 0, got %v", res)
	}

	res = sendCommand("HGET", "myhash", "field1")
	if res.Type != "bulk" || res.Bulk != "value2" {
		t.Errorf("Expected HGET to return 'value2', got %v", res)
	}

	res = sendCommand("HSET", "myhash", "field2", "hello")
	res = sendCommand("HGETALL", "myhash")
	if res.Type != "array" || len(res.Array) != 4 { // 2 fields, 2 values
		t.Errorf("Expected HGETALL to return array of 4, got %v", res)
	}

	res = sendCommand("HDEL", "myhash", "field1", "field2", "nonexistent")
	if res.Type != "integer" || res.Num != 2 {
		t.Errorf("Expected HDEL to return 2, got %v", res)
	}
}
