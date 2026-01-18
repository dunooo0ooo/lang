package bytecode

import "io"

type Chunk struct {
	Code      []byte
	Constants []Value
}

func (c *Chunk) Write(op OpCode) {
	c.Code = append(c.Code, byte(op))
}

func (c *Chunk) WriteByte(b byte) error {
	c.Code = append(c.Code, b)
	return nil
}

func (c *Chunk) WriteUint16(v uint16) {
	c.Code = append(c.Code, byte(v>>8), byte(v))
}

func (c *Chunk) PatchUint16(offset int, v uint16) error {
	if offset+1 >= len(c.Code) {
		return io.ErrShortWrite
	}
	c.Code[offset] = byte(v >> 8)
	c.Code[offset+1] = byte(v)
	return nil
}

func (c *Chunk) AddConstant(v Value) int {
	c.Constants = append(c.Constants, v)
	return len(c.Constants) - 1
}

// Дополнительные методы для удобства
func (c *Chunk) WriteInstruction(op OpCode, args ...byte) {
	c.Write(op)
	c.Code = append(c.Code, args...)
}

func (c *Chunk) GetCodeSize() int {
	return len(c.Code)
}

func (c *Chunk) GetConstantCount() int {
	return len(c.Constants)
}

func (c *Chunk) Clear() {
	c.Code = nil
	c.Constants = nil
}