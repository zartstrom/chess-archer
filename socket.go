package main

import (
    "github.com/gorilla/websocket"
)

func NewSocket(conn *websocket.Conn) *Socket {
    return &Socket{
        conn:   conn,
		output: make(chan string),
		input:  make(chan string),
		err:    make(chan bool),
    }

}

type Socket struct {
    // The websocket connection.
    conn   *websocket.Conn
    output chan string
    input  chan string
    err    chan bool
}

func (s *Socket) Output() chan string { return s.output }
func (s *Socket) Input() chan string { return s.input }
func (s *Socket) Err() chan bool { return s.err }

func (s *Socket) Start() {
    go s.writer()
    go s.reader()
}

func (s *Socket) Terminate() {
    s.conn.Close()
}

func (s *Socket) reader() {
    for {
        _, bytes, err := s.conn.ReadMessage()
        if err != nil {
            s.err <- true
            break
        }
        s.input <- string(bytes)
    }
    s.conn.Close()
}

func (s *Socket) writer() {
    for message := range s.output {
        bytes := []byte(message)
        err := s.conn.WriteMessage(websocket.TextMessage, bytes)
        if err != nil {
            break
        }
    }
    s.conn.Close()
}

