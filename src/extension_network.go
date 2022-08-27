package interpreter

import (
	"net"
	"strconv"
	"time"
)

// The network extension
type Network struct {
	// Universal timeout
	timeout time.Duration
	// Port in the ":42000" format
	port string
}

// Sets the timeout corresponding to the given byte
// Formula: timeout = 0.1 second * b
func (n *Network) SetTimeout(b byte) {
	n.timeout = (time.Second / 10) * time.Duration(b)
}

// Sets the port corresponding to the given byte
// Formula: port = 24000 + b
func (n *Network) SetPort(b byte) {
	n.port = byteToPort(b)
}

// Attempts to send a byte of data to localhost at the saved port
// Returns true if success, false otherwise
func (n *Network) Send(b byte) bool {
	conn, err := net.DialTimeout("tcp4", "127.0.0.1:"+n.port, n.timeout)
	if err != nil {
		return false
	}
	nBytes, err := conn.Write([]byte{b})
	if err != nil || nBytes != 1 {
		conn.Close()
		return false
	}
	conn.Close()
	return true
}

// Attempts to send a byte of data to localhost at the saved port
// Returns the received byte if successful, 0 otherwise
func (n *Network) Receive() byte {
	// Setup Connection
	listener, err := net.Listen("tcp4", "0.0.0.0:"+n.port)
	if err != nil {
		return 0
	}
	// Listen Async
	resultChan := make(chan byte, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			resultChan <- 0
			return
		}
		defer conn.Close()
		b := make([]byte, 1)
		nBytes, err := conn.Read(b)
		if nBytes == 0 || err != nil {
			resultChan <- 0
			return
		}
		resultChan <- b[0]
	}()
	// Wait on listener or timer
	select {
	case b := <-resultChan:
		listener.Close()
		return b
	case <-time.After(n.timeout):
		listener.Close()
		return 0
	}
}

// Returns network with default values
func NewNetwork() *Network {
	return &Network{
		timeout: 5 * time.Second,
		port:    ":42000",
	}
}

// Convert a byte to a port in the "####" format starting from 42000 to 42255
func byteToPort(b byte) string {
	return strconv.FormatInt(int64(b)+42000, 10)
}
