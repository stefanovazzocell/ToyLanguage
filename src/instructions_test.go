package interpreter

import (
	"bytes"
	"strings"
	"testing"
)

/*
* Tests
**/

func TestNewInstructions(t *testing.T) {
	testCases := []struct {
		code         string
		instructions []byte
		err          bool
	}{
		{"", []byte{}, false},
		{"tl:ten:this is an empty program", []byte{}, false},
		{"tl:nexx.+[.+]", []byte{byte('.'), byte('+'), byte('['), byte('.'), byte('+'), byte(']')}, false},
		{"tl:neet some .+ unexpected[.text\n+]", []byte{byte('.'), byte('+'), byte('['), byte('.'), byte('+'), byte(']')}, false},
	}

	for _, test := range testCases {
		i, err := NewInstructions(strings.NewReader(test.code))
		if i == nil {
			t.Fatal("Failed to setup instructions: null pointer")
		}
		if !test.err && err != nil {
			t.Fatalf("Expected no error for %q, got %v\n", test.code, err)
		}
		if test.err && err == nil {
			t.Fatalf("Expected error for %q, got nil\n", test.code)
		}
		if i.pc != 0 {
			t.Fatalf("PC counter initialized to %d, expected 0", i.pc)
		}
		if !bytes.Equal(i.instruction, test.instructions) {
			t.Fatalf("Expected %+d, Got %+d", test.instructions, i.instruction)
		}
		if i.extensions != 0 {
			t.Fatalf("Expected no extensions, instead found %b", i.extensions)
		}
	}
	i, err := NewInstructions(strings.NewReader("tl:net:xxx[]"))
	if i == nil {
		t.Fatal("Failed to setup instructions: null pointer")
	}
	if err != nil {
		t.Fatalf("Expected no error, instead got %v\n", err)
	}
	if i.extensions&ExtNet != ExtNet {
		t.Fatalf("Failed to parse extensions. Got %b", i.extensions)
	}
	if !bytes.Equal([]byte{'[', ']'}, i.instruction) {
		t.Fatalf("Failed to parse instruction. Got %v", i.instruction)
	}
}

func TestIsValidInstruction(t *testing.T) {
	validBytesCore := map[byte]bool{
		'>': true,
		'<': true,
		'+': true,
		'-': true,
		'.': true,
		',': true,
		'[': true,
		']': true,
	}

	validBytesNetwork := map[byte]bool{
		'*': true,
		'@': true,
		'?': true,
		'^': true,
	}

	for i := 0; i < 256; i++ {
		inst := byte(i)
		_, validBase := validBytesCore[inst]
		_, validNet := validBytesNetwork[inst]
		actualBase := IsValidInstruction(inst, 0)
		actualNet := IsValidInstruction(inst, ExtNet)
		actualAny := IsValidInstruction(inst, 0b11111111)
		if validBase && (!actualBase || !actualNet || !actualAny) {
			t.Errorf("Instruction %d is a valid base instruction but did not match correctly", inst)
		} else if !validBase && actualBase {
			t.Errorf("Instruction %d is not a valid base but it was reported as such", inst)
		}
		if validNet && (actualBase || !actualNet || !actualAny) {
			t.Errorf("Instruction %d is a valid net instruction but did not match correctly", inst)
		} else if (!validNet && !validBase) && (actualBase || actualNet) {
			t.Errorf("Instruction %d is not a valid net but it was reported as such", inst)
		}
	}
}

func TestJumping(t *testing.T) {
	testCasesForward := []struct {
		code             string
		startingPosition int
		startingChar     byte
		endingPosition   int
		endingChar       byte
		expectedReturn   bool
	}{
		{"", 0, '[', 0, ']', false},
		{"[]", 1, '[', 2, ']', true},
		{"[+++]", 1, '[', 5, ']', true},
		{"[[+]]", 1, '[', 5, ']', true},
		{"+[+.+]+", 2, '[', 6, ']', true},
		{"+[+.++", 2, '[', 6, ']', false},
		{"+]+.++", 6, '[', 6, ']', false},
	}

	for _, test := range testCasesForward {
		// Parse code
		i, err := NewInstructions(strings.NewReader(test.code))
		if err != nil {
			t.Fatalf("Failed to parse code: %v", err)
		}
		// Jump forward
		i.pc = test.startingPosition
		if i.JumpForward(test.endingChar) != test.expectedReturn {
			t.Fatalf("Expected %v as forward return for %q, instead got %v",
				test.expectedReturn, test.code, !test.expectedReturn)
		}
		if i.pc != test.endingPosition {
			t.Fatalf("Expected ending at %d, instead got %d for %q",
				test.endingPosition, i.pc, test.code)
		}
	}

	testCasesBackward := []struct {
		code             string
		endingPosition   int
		endingChar       byte
		startingPosition int
		startingChar     byte
		expectedReturn   bool
	}{
		{"", 0, '[', 0, ']', false},
		{"[]", 1, '[', 2, ']', true},
		{"[+++]", 1, '[', 5, ']', true},
		{"[[+]]", 1, '[', 5, ']', true},
		{"+[+.+]+", 2, '[', 6, ']', true},
		{"+[+.++", 0, '[', 6, ']', false},
		{"+]+.++", 0, '[', 2, ']', false},
	}
	for _, test := range testCasesBackward {
		// Parse code
		i, err := NewInstructions(strings.NewReader(test.code))
		if err != nil {
			t.Fatalf("Failed to parse code: %v", err)
		}
		// Jump forward
		i.pc = test.startingPosition
		if i.JumpBackward(test.endingChar) != test.expectedReturn {
			t.Fatalf("Expected %v as backward return for %q, instead got %v",
				test.expectedReturn, test.code, !test.expectedReturn)
		}
		if i.pc != test.endingPosition {
			t.Fatalf("Expected ending at %d, instead got %d for %q",
				test.endingPosition, i.pc, test.code)
		}
	}
}

func TestPop(t *testing.T) {
	i, err := NewInstructions(strings.NewReader(".+[.+]"))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}
	expectedBytes := []byte(".+[.+]")
	count := 1
	for _, expected := range expectedBytes {
		actual := i.Pop()
		if actual != expected {
			t.Fatalf("Expected %c, got %c", expected, actual)
		}
		if i.pc != count {
			t.Fatalf("PC is at %d instead of expected %d", i.pc, count)
		}
		count++
	}
	if i.pc != 6 {
		t.Fatalf("PC is at %d instead of terminal position", i.pc)
	}
	actual := i.Pop()
	if actual != 0 {
		t.Fatalf("Got %c instead of NULL at the end of string", actual)
	}
	actual = i.Pop()
	if actual != 0 {
		t.Fatalf("Got %c instead of NULL after second try at end of string", actual)
	}
	i.Reset()
	if i.pc != 0 {
		t.Fatalf("PC is at %d instead of reset (0) position", i.pc)
	}
}

func TestCountRepeating(t *testing.T) {
	i, err := NewInstructions(strings.NewReader("tl:net @+@@@+@"))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}
	if i.CountRepeating() != 0 {
		t.Fatalf("Expected 0 from CountRepeating() before start, got %d instead", i.CountRepeating())
	}
	if popped := i.Pop(); popped != '@' {
		t.Fatalf("Expected '@' from pop, got '%c' instead", popped)
	}
	if i.CountRepeating() != 1 {
		t.Fatalf("Expected 1 repeating instruction, got %d instead", i.CountRepeating())
	}
	if popped := i.Pop(); popped != '+' {
		t.Fatalf("Expected '+' from pop, got '%c' instead", popped)
	}
	if popped := i.Pop(); popped != '@' {
		t.Fatalf("Expected '@' from pop, got '%c' instead", popped)
	}
	if i.CountRepeating() != 3 {
		t.Fatalf("Expected 3 repeating instruction, got %d instead", i.CountRepeating())
	}
	if popped := i.Pop(); popped != '+' {
		t.Fatalf("Expected '+' from pop, got '%c' instead", popped)
	}
	if popped := i.Pop(); popped != '@' {
		t.Fatalf("Expected '@' from pop, got '%c' instead", popped)
	}
	if i.CountRepeating() != 1 {
		t.Fatalf("Expected 1 repeating instruction, got %d instead", i.CountRepeating())
	}
}

/*
* Benchmarks
**/

func BenchmarkNewInstructions(b *testing.B) {
	benchmarks := map[string]string{
		"HelloWorld":           "++++++++++[>+++++++>++++++++++>+++>+<<<<-]>++.>+.+++++++..+++.>++.<<+++++++++++++++.>.+++.------.--------.>+.>.",
		"DisplayAscii":         ".+[.+]",
		"DisplayAsciiPolluted": " some .+ unexpected[.text\n+]",
		"Fibonacci":            ">++++++++++>+>+[[+++++[>++++++++<-]>.<++++++[>--------<-]+<<<]>.>>[[-]<[>+<-]>>[<<+>+>-]<[>+<-[>+<-[>+<-[>+<-[>+<-[>+<-[>+<-[>+<-[>+<-[>[-]>+>+<<<-[>+<-]]]]]]]]]]]+>>>]<<<]",
	}

	for testName, testCode := range benchmarks {
		b.Run(testName, func(b *testing.B) {
			for i := b.N - 1; i >= 0; i-- {
				NewInstructions(strings.NewReader(testCode))
			}
		})
	}

}

func BenchmarkIsValidInstruction(b *testing.B) {
	for i := b.N - 1; i >= 0; i-- {
		// Allow all extensions
		IsValidInstruction(byte(i), 0b11111111)
	}
}
