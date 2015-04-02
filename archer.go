package main

import (
    "fmt"
    "github.com/gorilla/websocket"
    "io"
    "log"
    "net/http"
    "os/exec"
)

func check(e error) {
    if e != nil {
		panic(e)
	}
}

func engineAvailable(name string) bool {
    filepath, err := exec.Command("which", name).Output()
    if err != nil {
        fmt.Printf("Engine '%s' not availabe\n", name)
        return false
    } else {
        fmt.Println(string(filepath))
        return true
    }
}

type LaunchedProcess struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
}

func launchCmd(commandName string, commandArgs []string, env []string) (*LaunchedProcess, error) {
	cmd := exec.Command(commandName, commandArgs...)
	cmd.Env = env

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	return &LaunchedProcess{cmd, stdin, stdout, stderr}, err
}

func match(in, out chan string) {
    out <- "Waiting ..."

    // put elsewhere the process setup
    launched, _ := launchCmd("stockfish", []string{}, []string{})
    process := NewProcessEndpoint(launched)
	process.StartReading()
	defer process.Terminate()

    for {
        select {
        case engineTalk, ok := (<-process.Output()):
            out <- engineTalk
            log.Println(engineTalk)
            if !ok {
                return
            }
        case uciCommand := <-in:
            if !process.Send(uciCommand) {
                return
            }
        }
    }
}

type connection struct {
    // The websocket connection.
    ws *websocket.Conn

    send chan string
    receive chan string
}

func (c *connection) reader() {
    for {
        _, bytes, err := c.ws.ReadMessage()
        if err != nil {
            break
        }
        log.Println("Received: " + string(bytes))
        c.receive <- string(bytes)
    }
    c.ws.Close()
}

func (c *connection) writer() {
    for message := range c.send {
        bytes := []byte(message)
        err := c.ws.WriteMessage(websocket.TextMessage, bytes)
        if err != nil {
            break
        }
    }
    c.ws.Close()
}

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

func socketHandler(w http.ResponseWriter, r *http.Request) {
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    c := &connection{send: make(chan string), receive: make(chan string), ws: ws}
    go match(c.receive, c.send)
    go c.writer()
    c.reader()
}

func main() {
    http.Handle("/", http.FileServer(http.Dir(".")))
    http.HandleFunc("/socket", socketHandler)

    log.Println("serving")
    if err := http.ListenAndServe(":6400", nil); err != nil {
        log.Fatal("ListenAndServe:", err)
    }
}

