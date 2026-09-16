package main

import (
	"net"
	"testing"
	"time"

	"multi-threaded-Redis/Internal/resp"
)

func TestTCPServerPingPong(t *testing.T) {
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

	// We can use our writer and parser to communicate with our server
	writer := resp.NewWriter(conn)
	parser := resp.NewParser(conn)

	// Send PING command
	pingCmd := resp.Value{
		Type: "array",
		Array: []resp.Value{
			{Type: "bulk", Bulk: "PING"},
		},
	}

	err = writer.Write(pingCmd)
	if err != nil {
		t.Fatalf("Failed to write to server: %v", err)
	}

	// Read response
	response, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to read from server: %v", err)
	}

	// Verify response is +PONG\r\n
	if response.Type != "string" || response.Str != "PONG" {
		t.Errorf("Expected PONG, got %v", response)
	}
}
