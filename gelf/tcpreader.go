package gelf

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

type TCPReader struct {
	mu       sync.Mutex
	okToRead sync.Mutex
	listener *net.TCPListener
	conn     net.Conn
	cBuf     []byte
}

func newTCPReader(addr string) (*TCPReader, chan string, error) {
	var err error
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("ResolveTCPAddr('%s'): %s", addr, err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("ListenTCP: %s", err)
	}

	r := new(TCPReader)
	r.listener = listener
	r.okToRead.Lock()
	signal := make(chan string, 1)

	go r.listenUntilCloseSignal(signal)

	return r, signal, nil
}

func (r *TCPReader) listenUntilCloseSignal(signal chan string) {
	defer func() { signal <- "done" }()
	defer r.listener.Close()
	for {
		conn, err := r.listener.Accept()
		if err != nil {
			break
		}
		go r.handleConnection(conn)
		select {
		case sig := <-signal:
			if sig == "stop" {
				break
			}
		default:
		}
	}
}

func (r *TCPReader) addr() string {
	return r.listener.Addr().String()
}

func (r *TCPReader) handleConnection(conn net.Conn) {
	defer conn.Close()

	r.mu.Lock()
	defer r.mu.Unlock()
	r.cBuf = nil

	reader := bufio.NewReader(conn)
	buffer, err := reader.ReadBytes(0)
	if err == nil {
		r.cBuf = buffer
		r.okToRead.Unlock()
	}
}

func (r *TCPReader) readMessage() (*Message, error) {
	r.okToRead.Lock()
	r.mu.Lock()
	defer r.mu.Unlock()

	var cReader *bytes.Reader

	cReader = bytes.NewReader(r.cBuf)

	msg := new(Message)
	if err := json.NewDecoder(cReader).Decode(&msg); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %s", err)
	}

	r.cBuf = nil
	return msg, nil
}

func (r *TCPReader) Close() {
	r.listener.Close()
}
