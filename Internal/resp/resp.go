package resp

// Value represents a generic RESP value
type Value struct {
	Type  string
	Str   string
	Num   int
	Bulk  string
	Array []Value
}
