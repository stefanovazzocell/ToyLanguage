package interpreter

import (
	"errors"
	"net"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	// The target address
	NetTargetAddr = "127.0.0.1:"
	// The listening address
	NetListenAddr = "0.0.0.0:"
	// Long timeout used for timeout = 0
	// If the timeout is set to this all ops must retry until success
	NetLongTimeout = time.Minute
)

const (
	// Idling
	netStateIdle uint8 = 0
	// Listening for incoming connection
	netStateListening uint8 = 1
	// We have an active connection
	netStateConnected uint8 = 2
)

// The network extension
type Network struct {
	// Lock on this data structure
	isLocked atomic.Bool
	// Internal state
	state uint8
	// Universal timeout
	timeout time.Duration
	// Port in the "42000" format
	port string
	// Cached listener
	listener net.Listener
	// Cached connection
	conn net.Conn
	// Send Cache
	sendCache []byte
}

// Lock
func (n *Network) lock(lowPriority bool) {
	for !n.isLocked.CompareAndSwap(false, true) {
		if lowPriority {
			runtime.Gosched()
		}
	}
}

// Unlock
func (n *Network) unlock() {
	n.isLocked.Store(false)
}

// Closes all connections and resets the cache
// clearCache controls if the cache is reset or not
// NOTE: Requires the lock to be held by the current process
func (n *Network) reset(clearCache bool) {
	if n.state == netStateListening {
		n.listener.Close()
	}
	if n.state == netStateConnected {
		n.conn.Close()
	}
	if clearCache {
		n.sendCache = []byte{}
	}
	n.state = netStateIdle
}

// Sets up the listener
// Returns true on success, false otherwise
// NOTE: Requires the lock to be held by the current process
func (n *Network) startListening() bool {
	// Init
	n.reset(true)
	// Setup listener
	listener, err := net.Listen("tcp4", NetListenAddr+n.port)
	if err != nil {
		return false
	}
	// Accept connections
	cAccepting := make(chan bool)
	go func(listener net.Listener) {
		cAccepting <- true
		conn, err := listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			// This listener was closed
			return
		}
		listener.Close()
		if err == nil {
			n.lock(false)
			n.state = netStateConnected
			n.conn = conn
			n.listener = nil
			n.unlock()
		}
	}(listener)
	// If ready to accept, finalize setup and terminate
	<-cAccepting
	n.listener = listener
	n.state = netStateListening
	return true
}

// Setup a connection to the current target
// Returns true if successful, false otherwise
// NOTE: Requires the lock to be held by the current process
func (n *Network) setupConnection() bool {
	// Init (but don't reset the send cache)
	n.reset(false)
	// Connect
	conn, err := net.DialTimeout("tcp4", NetTargetAddr+n.port, n.timeout)
	if err != nil {
		return false
	}
	n.state = netStateConnected
	n.conn = conn
	return true
}

// Sets the timeout corresponding to the given byte
// Formula: timeout = 0.1 second * b
func (n *Network) SetTimeout(b byte) {
	n.lock(false)
	if b == 0 {
		// Special case: if timeout is 0, receive/send is blocking until success
		n.timeout = NetLongTimeout
		return
	}
	n.timeout = time.Duration(b) * (time.Second / 10)
	n.unlock()
}

// Sets the port corresponding to the given byte
// Formula: port = 24000 + b
func (n *Network) SetPort(b byte) {
	// Get lock
	n.lock(false)
	// Update port
	n.port = byteToPort(b)
	// Reset
	n.reset(true)
	// Unlock
	n.unlock()
}

// Send queue over the network
// Also removes the data from send cache if successful
// Assumes lock held by caller and conn is valid
func (n *Network) send() bool {
	for len(n.sendCache) > 0 {
		var pkt []byte
		if len(n.sendCache) > 1024 {
			pkt, n.sendCache = n.sendCache[0:1024], n.sendCache[1024:]
		} else {
			pkt, n.sendCache = n.sendCache, []byte{}
		}
		n.conn.SetDeadline(time.Now().Add(n.timeout))
		nSent, err := n.conn.Write(pkt)
		if nSent != len(pkt) {
			n.sendCache = append(pkt, n.sendCache...)
			return false
		}
		if err != nil {
			n.sendCache = append(pkt, n.sendCache...)
			return false
		}
	}
	return true
}

// Same as Push, but doesn't retry and assumes lock is held by caller
func (n *Network) pushOnce() bool {
	if n.state == netStateConnected && n.send() {
		// Used existing connection successful
		return true
	}
	if n.state != netStateIdle {
		// We're not idle, so reset
		n.reset(false)
	}
	if !n.setupConnection() {
		// We failed to setup a connection
		return false
	}
	return n.send()
}

// Attempts to send queued data to NetTargetAddr at the saved port
// Returns true if success, false otherwise
func (n *Network) Push() bool {
	n.lock(false)
	defer n.unlock()
	for {
		if n.pushOnce() {
			return true
		}
		if n.timeout != NetLongTimeout {
			// Not a blocking call, return false
			return false
		}
	}
}

// Adds data to the send queue
func (n *Network) QueueSend(b byte) {
	n.lock(false)
	n.sendCache = append(n.sendCache, b)
	n.unlock()
}

// Internal version of receive
// Requires caller to held lock and only tries once
// Returns received data or false
func (n *Network) receiveOnce() (byte, bool) {
	singleByte := make([]byte, 1)
	if n.state == netStateConnected {
		// Try receiving now
		n.conn.SetDeadline(time.Now().Add(n.timeout))
		nSent, err := n.conn.Read(singleByte)
		if nSent == 1 && err == nil {
			return singleByte[0], true
		}
	}
	if n.state != netStateIdle && n.state != netStateListening {
		n.reset(true)
	}
	if n.state != netStateListening && !n.startListening() {
		return 0, false
	}
	timedOut := atomic.Bool{}
	time.AfterFunc(n.timeout, func() { timedOut.Store(true) })
	n.unlock()
	for !timedOut.Load() {
		n.lock(false)
		if n.state == netStateIdle {
			return 0, false
		}
		if n.state == netStateConnected {
			n.conn.SetDeadline(time.Now().Add(n.timeout))
			nSent, err := n.conn.Read(singleByte)
			if nSent == 1 && err == nil {
				return singleByte[0], true
			}
			return 0, false
		}
		n.unlock()
		time.Sleep(time.Second / 100)
	}
	return 0, false
}

// Attempts to receive a byte of data from NetListenAddr at the saved port
// Returns the received byte if successful, 0 otherwise
func (n *Network) Receive() byte {
	n.lock(false)
	defer n.unlock()
	for {
		b, ok := n.receiveOnce()
		if ok {
			return b
		}
		if n.timeout != NetLongTimeout {
			// Not a blocking call, return 0
			return 0
		}
	}
}

// Returns network with default values
func NewNetwork() *Network {
	return &Network{
		isLocked:  atomic.Bool{},
		state:     netStateIdle,
		timeout:   5 * time.Second,
		port:      ":42000",
		listener:  nil,
		conn:      nil,
		sendCache: []byte{},
	}
}

// Convert a byte to a port in the "####" format starting from 42000 to 42255
func byteToPort(b byte) string {
	return strconv.FormatInt(int64(b)+42000, 10)
}
