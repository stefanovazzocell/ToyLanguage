package interpreter

import (
	"bytes"
	"io"
)

type ExtensionCode uint8

var (
	ExtNet ExtensionCode = 0b00000001
)

var SupportedExtensions = map[string]ExtensionCode{
	"net": ExtNet,
}

// Structure containing the instructions and a program counter
type Instructions struct {
	instruction []byte        // Our instructions represented in ASCII bytes
	extensions  ExtensionCode // Enabled extensions
	pc          int           // Program Coutner
}

// Resets the PC
func (i *Instructions) Reset() {
	i.pc = 0
}

// Get the current instruction, increment the counter
// returns the current instruction or 0 (terminate program)
func (i *Instructions) Pop() byte {
	// Special case
	if i.pc == len(i.instruction) {
		return 0
	}
	// Return and increment
	i.pc++
	return i.instruction[i.pc-1]
}

// Moves the pointer forward until we reach the matching end byte
// Returns true if successful, false if not found
func (i *Instructions) JumpForward(matching byte) bool {
	if i.pc == 0 || i.pc == len(i.instruction) {
		// The program already terminated (or the request is invalid)
		return false
	}
	opened := 0
	// The last executed instruction
	initial := i.instruction[i.pc-1]
	for {
		if i.instruction[i.pc] == matching {
			if opened == 0 {
				// Found it!
				i.pc++
				return true
			} else {
				// It's not the matching entry
				opened--
			}
		} else if i.instruction[i.pc] == initial {
			// We found another opening
			opened++
		}
		i.pc++
		if i.pc == len(i.instruction) {
			// We ran out of instructions
			return false
		}
	}
}

// Moves the pointer backward until we reach the matching end byte
// Returns true if successful, false if not found
func (i *Instructions) JumpBackward(matching byte) bool {
	if i.pc <= 1 {
		// The program already at beginning
		return false
	}
	closed := 0
	i.pc--
	// The last executed instruction
	final := i.instruction[i.pc]
	i.pc--
	for {
		if i.pc == -1 {
			// We ran out of instructions
			i.pc = 0
			return false
		}
		if i.instruction[i.pc] == matching {
			if closed == 0 {
				// Found it!
				i.pc++
				return true
			} else {
				// It's not the matching entry
				closed--
			}
		} else if i.instruction[i.pc] == final {
			// We found another closing
			closed++
		}
		i.pc--
	}
}

// Returns the number of times the last instruction repeats consecutevely
// Also advances the program counter accordingly
// (including the original appearance)
// Special case: returns 0 if invalid request
func (i *Instructions) CountRepeating() int {
	if i.pc == 0 {
		// Invalid request: program hasn't executed any instruction yet
		return 0
	}
	previous := i.instruction[i.pc-1]
	count := 1
	for p := i.pc; p < len(i.instruction); p++ {
		if i.instruction[p] != previous {
			i.pc += (count - 1)
			return count
		}
		count++
	}
	return count
}

// Parse the instructions from a given io reader.
// Returns Instructions and an error
func NewInstructions(reader io.Reader) (*Instructions, error) {
	// Setup
	instructions := Instructions{
		instruction: []byte{},
		extensions:  0,
		pc:          0,
	}
	inst := []byte{}
	// Load data from our reader
	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		inst = append(inst, buf[:n]...)
		if err == io.EOF {
			break
		}
		if err != nil {
			return &instructions, err
		}
	}
	// Check for extensions "tl:"
	// Supported extensions are: "net"
	// Fail quietly to improve compatibility with bf
	if len(inst) > 6 && inst[0] == 't' && inst[1] == 'l' && inst[2] == ':' {
		ext := make([]byte, 0, 3)
		for i := 3; i < len(inst); i++ {
			if inst[i] == ':' {
				extCode, supported := SupportedExtensions[string(ext)]
				if !supported {
					// Not supported, stop parsing
					break
				}
				instructions.extensions |= extCode
				ext = make([]byte, 0, 3)
				continue
			}
			if len(ext) == 3 {
				extCode, supported := SupportedExtensions[string(ext)]
				if !supported {
					// Not supported, stop parsing
					break
				}
				instructions.extensions |= extCode
				// Anything else is too long to be a valid extension
				break
			}
			if bytes.IndexByte([]byte("net"), inst[i]) == -1 {
				// Not a valid char
				break
			}
			ext = append(ext, inst[i])
		}
	}
	// Filter valid instructions
	n := 0
	for i := 0; i < len(inst); i++ {
		if IsValidInstruction(inst[i], instructions.extensions) {
			inst[n] = inst[i]
			n++
		}
	}
	instructions.instruction = inst[:n]
	// Return
	return &instructions, nil
}

// Returns true if b is a valid instruction, false otherwise
// ext represents the enabled extensions, all non-compliant bytes will be ignored
func IsValidInstruction(b byte, ext ExtensionCode) bool {
	return (b == byte('>') || b == byte('<') || // Base: Move Pointer
		b == byte('+') || b == byte('-') || // Base: Incr/Decr Byte
		b == byte('.') || b == byte(',') || // Base: Write/Read Input
		b == byte('[') || b == byte(']') || // Base: Conditional Loop
		(ext&ExtNet == ExtNet) && // Extension: Network
			(b == byte('?') || b == byte('^') || b == byte('@') || b == byte('*') || b == byte(';')))
}
