package interpreter

import "errors"

var (
	ErrMemOutOfBoundary = errors.New("this operation tried moving the pointer out of boundary")
)

const (
	// Brainfuck has 30000 memory cells, this superset uses 2^16
	MemSize = 65536
)

// The program working memory
type Memory struct {
	mem [65536]byte // Memory
	p   int         // Memory pointer
}

// Blanks out the working memory and resets the pointer
func (m *Memory) Reset() {
	m.mem = [MemSize]byte{}
	m.p = 0
}

// Increases the current memory value
func (m *Memory) Incr() {
	m.mem[m.p]++
}

// Decreases the current memory value
func (m *Memory) Decr() {
	m.mem[m.p]--
}

// Returns the current byte
func (m *Memory) Get() byte {
	return m.mem[m.p]
}

// Sets b to the current byte
func (m *Memory) Set(b byte) {
	m.mem[m.p] = b
}

// Moves the pointer to the next value if possible
// Returns an error
func (m *Memory) Next() error {
	if m.p == MemSize-1 {
		return ErrMemOutOfBoundary
	}
	m.p++
	return nil
}

// Moves the pointer to the previous value if possible
// Returns an error
func (m *Memory) Prev() error {
	if m.p == 0 {
		return ErrMemOutOfBoundary
	}
	m.p--
	return nil
}

// Returns a copy of the memory from 0 to the last non-zero index
func (m *Memory) Bytes() []byte {
	// Find the last non-zero value
	lastNonZero := 0
	for i := MemSize - 1; i >= 0; i-- {
		if i == 0 && m.mem[i] == 0 {
			// Special case
			return []byte{}
		}
		if m.mem[i] != 0 {
			lastNonZero = i
			break
		}
	}
	// Collect all the bytes
	memBytes := make([]byte, lastNonZero+1)
	for i := lastNonZero; i >= 0; i-- {
		memBytes[i] = m.mem[i]
	}
	return memBytes
}

// Returns a blank memory
func NewMemory() *Memory {
	return &Memory{
		mem: [MemSize]byte{},
		p:   0,
	}
}
