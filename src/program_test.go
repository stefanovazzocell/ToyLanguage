package interpreter

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

/*
* Tests
**/

func TestProgram(t *testing.T) {
	p, err := NewProgram(strings.NewReader(".+[.+]"))
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}
	if len(p.Instructions.instruction) != 6 {
		t.Fatal("Failed to load instructions corretly in a new program")
	}
	p.Memory.Set(11)
	p.Memory.p = 2
	p.Memory.Set(17)
	p.Instructions.pc = 3
	p.Reset()
	if p.Memory.p != 0 || p.Instructions.pc != 0 || len(p.Memory.Bytes()) != 0 {
		t.Fatalf("Failed to reset program: %d, %d, %+d", p.Memory.p, p.Instructions.pc, p.Memory.Bytes())
	}
	if p.HasExtensions(ExtNet) {
		t.Fatal("Incorrectly reported Network Extension as enabled")
	}
	p.LoadProgram(strings.NewReader("tl:neet"))
	if len(p.Instructions.instruction) != 0 {
		t.Fatal("Failed to load an empty program")
	}
	if p.HasExtensions(ExtNet) {
		t.Fatal("Incorrectly reported Network Extension as enabled")
	}
	if len(p.GetInstructions()) != 0 {
		t.Fatalf("Expected no instructions, instead got %s", p.GetInstructions())
	}
}

func TestRunNext(t *testing.T) {
	t.Run("MemoryOps", func(t *testing.T) {
		p, err := NewProgram(strings.NewReader(`+ > +++ > +++ < -`))
		if err != nil {
			t.Fatalf("Failed to load test program: %v", err)
		}
		t.Logf("Parsed %s", string(p.Instructions.instruction))
		if err = p.Run(11); err != ErrExecutionLimit {
			t.Fatalf("Expected ErrExecutionLimit on Run, instead got %v", err)
		}
		if err = p.RunNext(); err != ErrProgramDone {
			t.Fatalf("Expected ErrProgramDone on RunNext, instead got %v", err)
		}
		t.Logf("PC: %d, MP: %d", p.Instructions.pc, p.Memory.p)
		actualMem := p.Memory.Bytes()
		expectedMem := []byte{1, 2, 3}
		if !bytes.Equal(actualMem, expectedMem) {
			t.Fatalf("Expected %+d, but memory is %+d", expectedMem, actualMem)
		}
	})
	t.Run("Loops", func(t *testing.T) {
		p, err := NewProgram(strings.NewReader(`+[>+<+] do nothing next >>>[+]`))
		if err != nil {
			t.Fatalf("Failed to load test program: %v", err)
		}
		t.Logf("Parsed %s", string(p.Instructions.instruction))
		if err = p.Run(2 + 5*256 + 4 + 1); err != nil {
			t.Fatalf("Expected no error on Run, instead got %v", err)
		}
		t.Logf("PC: %d, MP: %d", p.Instructions.pc, p.Memory.p)
		actualMem := p.Memory.Bytes()
		expectedMem := []byte{0, 255}
		if !bytes.Equal(actualMem, expectedMem) {
			t.Fatalf("Expected %+d, but memory is %s", expectedMem, actualMem)
		}
	})
	t.Run("IO", func(t *testing.T) {
		p, err := NewProgram(strings.NewReader(`@.+[.+] print all ASCII and write stdin in next cell >,`))
		if err != nil {
			t.Fatalf("Failed to load test program: %v", err)
		}
		t.Logf("Parsed %s", string(p.Instructions.instruction))
		// Custom writer/reader
		p.IOReader = strings.NewReader("!")
		outputBuffer := bytes.NewBuffer(make([]byte, 0, 256))
		p.IOWriter = outputBuffer
		// Run test
		if err = p.Run(1 + 3 + 3*255 + 2 + 1); err != nil {
			t.Fatalf("Expected no error on Run, instead got %v", err)
		}
		t.Logf("PC: %d, MP: %d", p.Instructions.pc, p.Memory.p)
		actualMem := p.Memory.Bytes()
		expectedMem := []byte{0, '!'}
		if !bytes.Equal(actualMem, expectedMem) {
			t.Fatalf("Expected %+d, but memory is %+d", expectedMem, actualMem)
		}
		expectedOutput := make([]byte, 256)
		for i := 0; i < 256; i++ {
			expectedOutput[i] = byte(i)
		}
		actualOutput := outputBuffer.Bytes()
		if !bytes.Equal(actualOutput, expectedOutput) {
			t.Fatalf("Expected 256 ASCII, instead got %+d", actualOutput)
		}
	})
	t.Run("Network", func(t *testing.T) {
		p, err := NewProgram(strings.NewReader(`tl:net ++*-@^?`))
		if err != nil {
			t.Fatalf("Failed to load test program: %v", err)
		}
		t.Logf("Parsed %q", string(p.Instructions.instruction))
		for i := 0; i < 3; i++ {
			if err := p.RunNext(); err != nil {
				t.Fatalf("Expected no error, instead got %v for %d's instruction", err, i)
			}
		}
		if p.Network.timeout != time.Millisecond*200 {
			t.Fatalf("The timeout wasn't set correctly: %q", p.Network.timeout.String())
		}
		for i := 0; i < 2; i++ {
			if err := p.RunNext(); err != nil {
				t.Fatalf("Expected no error, instead got %v for %d's instruction", err, i)
			}
		}
		if p.Network.port != "42001" {
			t.Fatalf("The port wasn't set correctly: %q", p.Network.port)
		}
		if err := p.RunNext(); err != nil {
			t.Fatalf("Expected no error, instead got %v for last instruction", err)
		}
		if p.Memory.Get() != 1 {
			t.Fatalf("Expected the send to fail and leave a non zero value in memory, instead got %d", p.Memory.Get())
		}
		if err := p.RunNext(); err != nil {
			t.Fatalf("Expected no error, instead got %v for last instruction", err)
		}
		if p.Memory.Get() != 0 {
			t.Fatalf("Expected the receive to fail and leave a zero value in memory, instead got %d", p.Memory.Get())
		}
		if err := p.RunNext(); err != ErrProgramDone {
			t.Fatalf("Expected ErrProgramDone, instead got %v for last instruction", err)
		}
	})
}

/*
* Benchmarks
**/

func BenchmarkInfiniteLoop(b *testing.B) {
	b.Run("RunNext", func(b *testing.B) {
		p, err := NewProgram(strings.NewReader(`+[]`))
		if err != nil {
			b.Fatalf("Failed to load test program: %v", err)
		}
		b.ResetTimer()
		for i := b.N; i > 0; i-- {
			p.RunNext()
		}
	})
	b.Run("RunWithLimit", func(b *testing.B) {
		p, err := NewProgram(strings.NewReader(`+[]`))
		if err != nil {
			b.Fatalf("Failed to load test program: %v", err)
		}
		b.ResetTimer()
		err = p.Run(b.N)
		b.StopTimer()
		if err != ErrExecutionLimit {
			b.Fatalf("Expected hitting executing limit, instead got %v", err)
		}
	})
}
