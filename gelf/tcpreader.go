package gelf

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type TCPReader struct {
	listener *net.TCPListener
	conn     net.Conn
	messages chan []byte
}

type conn_channels struct {
	drop    chan string
	confirm chan string
}

func newTCPReader(addr string) (*TCPReader, chan string, chan string, error) {
	var err error
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("ResolveTCPAddr('%s'): %s", addr, err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("ListenTCP: %s", err)
	}

	r := &TCPReader{
		listener: listener,
		messages: make(chan []byte, 100), // Make a buffered channel with at most 100 messages
	}

	close_signal := make(chan string, 1)
	done_signal := make(chan string, 1)

	go r.listenUntilCloseSignal(close_signal, done_signal)

	return r, close_signal, done_signal, nil
}

func (r *TCPReader) accepter(connections chan net.Conn) {
	for {
		conn, err := r.listener.Accept()
		if err != nil {
			break
		}
		connections <- conn
	}
}

func (r *TCPReader) listenUntilCloseSignal(close_signal chan string, done_signal chan string) {
	defer func() { done_signal <- "done" }()
	defer r.listener.Close()
	var conns []conn_channels
	connections_channel := make(chan net.Conn, 1)
	go r.accepter(connections_channel)
	for {
		select {
		case conn := <-connections_channel:
			drop_signal := make(chan string, 1)
			drop_confirm := make(chan string, 1)
			channels := conn_channels{drop: drop_signal, confirm: drop_confirm}
			go handleConnection(conn, r.messages, drop_signal, drop_confirm)
			conns = append(conns, channels)
		default:
		}

		select {
		case sig := <-close_signal:
			if sig == "stop" {
				if len(conns) >= 1 {
					for _, s := range conns {
						if s.drop != nil {
							s.drop <- "drop"
							<-s.confirm
							conns = append(conns[:0], conns[1:]...)
						}
					}
					return
				} else {
					close_signal <- "stop"
				}
			}
			if sig == "drop" {
				if len(conns) >= 1 {
					for _, s := range conns {
						if s.drop != nil {
							s.drop <- "drop"
							<-s.confirm
							conns = append(conns[:0], conns[1:]...)
						}
					}
				}
				done_signal <- "done"
			}
		default:
		}
	}
}

func (r *TCPReader) addr() string {
	return r.listener.Addr().String()
}

func handleConnection(conn net.Conn, messages chan<- []byte, drop_signal chan string, drop_confirm chan string) {
	defer func() { drop_confirm <- "done" }()
	defer conn.Close()
	reader := bufio.NewReader(conn)

	var b []byte
	var err error
	drop := false
	can_drop := false

	for {
		conn.SetDeadline(time.Now().Add(2 * time.Second))
		if b, err = reader.ReadBytes(0); err != nil {
			if drop {
				return
			}
		} else if len(b) > 0 {
			messages <- b
			can_drop = true
			if drop {
				return
			}
		} else if drop {
			return
		}
		select {
		case sig := <-drop_signal:
			if sig == "drop" {
				drop = true
				time.Sleep(1 * time.Second)
				if can_drop {
					return
				}
			}
		default:
		}
	}
}

func (r *TCPReader) readMessage() (*Message, error) {
	b := <-r.messages

	var msg Message
	if err := json.Unmarshal(b[:len(b)-1], &msg); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %s", err)
	}

	return &msg, nil
}

func (r *TCPReader) Close() {
	r.listener.Close()
}
