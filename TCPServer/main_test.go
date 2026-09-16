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
