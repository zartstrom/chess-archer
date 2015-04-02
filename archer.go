package main

import (
    "bufio"
    "fmt"
    "io"
    "os/exec"
    //"strings"
    "time"
    //"net/http"
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

func scanLineWise(r io.Reader) {
    fmt.Println("in scanLineWise")
	s := bufio.NewScanner(r)
	s.Split(bufio.ScanLines)

	for s.Scan() {
        fmt.Printf("%s\n", s.Text())
	}
}

func uciCommand(str string) []byte {
    return []byte(fmt.Sprintf("%s\n", str))
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

func callStockfish(commands chan string) {

    // TODO: Talk to stockfish

    if !engineAvailable("stockfish") { return }
    cmd := exec.Command("stockfish")

    stdout, err := cmd.StdoutPipe()
    check(err)
    //defer stdout.Close()

    stdin, err := cmd.StdinPipe()
    check(err)
    //defer stdin.Close()
    go cmd.Start()
    //defer cmd.Wait()

    go scanLineWise(stdout)

    go func() {
        fmt.Printf("go func in callStockfish\n")
        for clientCommand := range commands {
            fmt.Printf("clientCommand: %s\n", clientCommand)
            stdin.Write(uciCommand(clientCommand))
        }
        //close(out)
    }()

}

//func handler(w http.ResponseWriter, r *http.Request) {
//    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
//}

func PipeEndpoints(process Endpoint, ch chan string) {
	process.StartReading()
	//e2.StartReading()

	defer process.Terminate()
	//defer e2.Terminate()
	for {
		select {
		case msgOne, ok := (<-process.Output()):
            fmt.Println(msgOne)
			if !ok {
				return
			}
		case command := <-ch:
			if !process.Send(command) {
				return
			}
		}
	}
}

func main() {
    //http.HandleFunc("/", handler)
    //http.ListenAndServe(":6400", nil)
    //in := make(chan string)
    //go callStockfish(in)
    //in <- "uci"
    //in <- "go"
    //close(in)
    launched, _ := launchCmd("stockfish", []string{}, []string{})
    fmt.Printf("%d\n", launched.cmd.Process.Pid)
	//process := NewProcessEndpoint(launched, log)
	process := NewProcessEndpoint(launched)
    commandChannel := make(chan string)
    go PipeEndpoints(process, commandChannel)
    commandChannel <- "uci"
    commandChannel <- "go infinite"

    time.Sleep(2500 * time.Millisecond)

    //stdin.Write([]byte("uci\n"))
    //stdin.Write([]byte("go infinite\n"))
    //time.Sleep(2500 * time.Millisecond)
    //stdi.Write(uciCommand("stop"))
    //stdin.Write(uciCommand("go"))
}

