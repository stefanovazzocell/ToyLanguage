package interpreter

import (
	"fmt"
	"testing"
	"time"
)

/*
* Tests
**/

func TestNetwork(t *testing.T) {
	n := NewNetwork()
	if n == nil {
		t.Fatal("Failed to setup network: null pointer")
	}
	if n.port != ":42000" || n.timeout != time.Second*5 {
		t.Fatalf("Unexpected default values: port:%q, timeout:%s", n.port, n.timeout.String())
	}
	n.SetPort(1)
	n.SetPort(11)
	if n.port != "42011" {
		t.Fatalf("Port was not set correctly: %q", n.port)
	}
	n.SetTimeout(23)
	if n.timeout != time.Millisecond*2300 {
		t.Fatalf("Timeout was not set correctly: %s", n.timeout.String())
	}
}

func TestByteToPort(t *testing.T) {
	for i := 0; i < 256; i++ {
		actual := byteToPort(byte(i))
		expected := fmt.Sprintf("%d", 42000+i)
		if len(actual) != 5 || actual != expected {
			t.Errorf("Got %q but expected %q", actual, expected)
		}
	}
}

/*
* Benchmarks
**/

func BenchmarkByteToPort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		byteToPort(byte(i))
	}
}
