package aof

import (
	"bufio"
	"io"
	"multi-threaded-Redis/Internal/resp"
	"os"
	"sync"
)

type Aof struct {
	file *os.File
	mu   sync.Mutex
}

func NewAof(path string) (*Aof, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Aof{
		file: f,
	}, nil
}

func (a *Aof) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.file.Close()
}

func (a *Aof) Write(value resp.Value) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	writer := resp.NewWriter(a.file)
	return writer.Write(value)
}

func (a *Aof) Read(callback func(value resp.Value)) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Ensure we read from the beginning
	_, err := a.file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(a.file)
	parser := resp.NewParser(reader)

	for {
		val, err := parser.Parse()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		callback(val)
	}

	// Seek back to the end for future appends
	_, err = a.file.Seek(0, io.SeekEnd)
	return err
}
