package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type Parser struct {
	reader *bufio.Reader
}

func NewParser(rd io.Reader) *Parser {
	return &Parser{reader: bufio.NewReader(rd)}
}

func (p *Parser) Parse() (Value, error) {
	_type, err := p.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch _type {
	case '+':
		return p.parseSimpleString()
	case '-':
		return p.parseError()
	case ':':
		return p.parseInteger()
	case '$':
		return p.parseBulkString()
	case '*':
		return p.parseArray()
	default:
		return Value{}, fmt.Errorf("unknown type: %v", string(_type))
	}
}

func (p *Parser) readLine() (line []byte, n int, err error) {
	for {
		b, err := p.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		n += 1
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' && line[len(line)-1] == '\n' {
			break
		}
	}
	return line[:len(line)-2], n, nil
}

func (p *Parser) parseSimpleString() (Value, error) {
	v, _, err := p.readLine()
	if err != nil {
		return Value{}, err
	}
	return Value{Type: "string", Str: string(v)}, nil
}

func (p *Parser) parseError() (Value, error) {
	v, _, err := p.readLine()
	if err != nil {
		return Value{}, err
	}
	return Value{Type: "error", Str: string(v)}, nil
}

func (p *Parser) parseInteger() (Value, error) {
	v, _, err := p.readLine()
	if err != nil {
		return Value{}, err
	}
	i, err := strconv.Atoi(string(v))
	if err != nil {
		return Value{}, err
	}
	return Value{Type: "integer", Num: i}, nil
}

func (p *Parser) parseBulkString() (Value, error) {
	lenLine, _, err := p.readLine()
	if err != nil {
		return Value{}, err
	}

	bulkLen, err := strconv.Atoi(string(lenLine))
	if err != nil {
		return Value{}, err
	}

	if bulkLen == -1 {
		return Value{Type: "null"}, nil
	}

	bulk := make([]byte, bulkLen)
	_, err = io.ReadFull(p.reader, bulk)
	if err != nil {
		return Value{}, err
	}

	// Read trailing \r\n
	_, _, err = p.readLine()
	if err != nil {
		return Value{}, err
	}

	return Value{Type: "bulk", Bulk: string(bulk)}, nil
}

func (p *Parser) parseArray() (Value, error) {
	lenLine, _, err := p.readLine()
	if err != nil {
		return Value{}, err
	}

	arrayLen, err := strconv.Atoi(string(lenLine))
	if err != nil {
		return Value{}, err
	}

	if arrayLen == -1 {
		return Value{Type: "null"}, nil
	}

	array := make([]Value, arrayLen)
	for i := 0; i < arrayLen; i++ {
		val, err := p.Parse()
		if err != nil {
			return Value{}, err
		}
		array[i] = val
	}

	return Value{Type: "array", Array: array}, nil
}
