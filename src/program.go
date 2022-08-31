package interpreter

import (
	"errors"
	"io"
	"os"
)

var (
	ErrProgramDone    = errors.New("the program has terminated")
	ErrProgramUnknown = errors.New("received unknown instruction")
	ErrExecutionLimit = errors.New("reached the execution limit")
	ErrIoNoInput      = errors.New("failed to read input")
	ErrIoNoOutput     = errors.New("failed to write output")
	ErrNotImplemented = errors.New("this feature is not implemented")
)

type Program struct {
	// Program Instructions
	Instructions *Instructions
	// Program Memory
	Memory *Memory

	// Network Extension
	Network *Network

	// Writer for IO output
	IOWriter io.Writer
	// Reader for IO input
	IOReader io.Reader
}

// Runs the entire program until done, error, or reached execution limit
func (p *Program) Run(limit int) error {
	for i := 0; i < limit; i++ {
		err := p.RunNext()
		if err == ErrProgramDone {
			return nil
		}
		if err != nil {
			return err
		}
	}
	return ErrExecutionLimit
}

// Runs the next instruction
// Returns an error if any
func (p *Program) RunNext() error {
	instruction := p.Instructions.Pop()
	if instruction == 0 {
		return ErrProgramDone
	}
	/*
	* Base
	**/
	// Increment the data pointer
	if instruction == '>' {
		return p.Memory.Next()
	}
	// Decrement the data pointer
	if instruction == '<' {
		return p.Memory.Prev()
	}
	// Increment (by one) the byte at the data pointer
	if instruction == '+' {
		p.Memory.Incr()
		return nil
	}
	// Decrement (by one) the byte at the data pointer
	if instruction == '-' {
		p.Memory.Decr()
		return nil
	}
	// Output the value at the data pointer
	if instruction == '.' {
		n, err := p.IOWriter.Write([]byte{p.Memory.Get()})
		if n != 1 || err != nil {
			return ErrIoNoOutput
		}
		return nil
	}
	// Accept one byte of input, store it at the data pointer
	if instruction == ',' {
		b := make([]byte, 1)
		n, err := p.IOReader.Read(b)
		if n != 1 || err != nil {
			return ErrIoNoInput
		}
		p.Memory.Set(b[0])
		return nil
	}
	// If the data pointer byte is zero, jump to the next corresponding `]`
	if instruction == '[' {
		if p.Memory.Get() == 0 {
			p.Instructions.JumpForward(']')
		}
		return nil
	}
	// If the data pointer byte is non-zero, jump to the previous corresponding `[`
	if instruction == ']' {
		if p.Memory.Get() != 0 {
			p.Instructions.JumpBackward('[')
		}
		return nil
	}
	/*
	* Extension: Network
	**/
	extNet := p.Instructions.extensions&ExtNet == ExtNet
	// Sets the timeout to the byte at the data pointer times 0.1 seconds
	if instruction == '*' && extNet {
		p.Network.SetTimeout(p.Memory.Get())
		return nil
	}
	// Set the port based on the byte at the data pointer
	if instruction == '@' && extNet {
		p.Network.SetPort(p.Memory.Get())
		return nil
	}
	// Listen for a message and writes the received value to the byte at the data pointer
	// On error sets the byte at the data pointer to `0`
	if instruction == '?' && extNet {
		p.Memory.Set(p.Network.Receive())
		return nil
	}
	// Queues the byte at the data pointer to be sent
	if instruction == '^' && extNet {
		p.Network.QueueSend(p.Memory.Get())
		return nil
	}
	// Sends the queued data to the target port
	// Sets the data pointer value to `0` is successful
	if instruction == ';' && extNet {
		if ok := p.Network.Push(); ok {
			p.Memory.Set(0)
		}
		return nil
	}

	return ErrProgramUnknown
}

// Rests memory and the program counter
func (p *Program) Reset() {
	p.Instructions.Reset()
	p.Memory.Reset()
}

// Loads a new program (without resetting memory)
func (p *Program) LoadProgram(r io.Reader) error {
	inst, err := NewInstructions(r)
	if err != nil {
		return err
	}

	p.Instructions = inst
	return nil
}

// Returns true if the given extension is enabled
func (p Program) HasExtensions(ec ExtensionCode) bool {
	return p.Instructions.extensions&ec == ec
}

// Returns the parsed instructions
func (p Program) GetInstructions() []byte {
	return p.Instructions.instruction
}

// Returns a new empty program
func NewProgram(r io.Reader) (Program, error) {
	inst, err := NewInstructions(r)
	if err != nil {
		return Program{}, err
	}

	return Program{
		Instructions: inst,
		Memory:       NewMemory(),
		Network:      NewNetwork(),
		IOWriter:     os.Stdout,
		IOReader:     os.Stdin,
	}, nil
}
