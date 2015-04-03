package main

import (
    //"fmt"
    "github.com/gorilla/websocket"
    "log"
    "net/http"
)

type Wire interface {
    Output() chan string
    Input() chan string
    Err() chan bool
    Start()
    Terminate()
}

func directedPlug(w1, w2 Wire) {
    // from source to consumer
    // the engine is the source of data
    // the socket steers the engine by command
    // need a good name
    // w1 = engine, w2 = socket
    w1.Start()
    w2.Start()
    defer w1.Terminate()
    defer w2.Terminate()

    for {
        select {
        case msg := <-w1.Output():
            log.Println(msg)
            w2.Output() <- msg

        case msg := <-w2.Input():
            log.Println(msg)
            w1.Input() <- msg

        case <-w1.Err():
            return
        case <-w2.Err():
            return
        }
    }
}

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

func socketHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    soc := NewSocket(conn)
    eng := NewEngine("stockfish")
    directedPlug(eng, soc)
}

func main() {
    http.Handle("/", http.FileServer(http.Dir(".")))
    http.HandleFunc("/socket", socketHandler)

    log.Println("serving")
    if err := http.ListenAndServe(":6400", nil); err != nil {
        log.Fatal("ListenAndServe:", err)
    }
}

