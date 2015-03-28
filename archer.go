package main

import "bufio"
import "fmt"
import "os"
import "os/exec"


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


func main() {

    // TODO: Talk to stockfish
    //start a stockfish instance
    //send command ("uci")
    //capture output

    if !engineAvailable("stockfish") { return }

    cmd := exec.Command("stockfish")

    stdout, err := cmd.StdoutPipe()
    check(err)

    stdin, err := cmd.StdinPipe()
    check(err)
    defer stdin.Close()

    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    cmd.Start()
    defer cmd.Wait()

    stdin.Write([]byte("uci"))

    r := bufio.NewReader(stdout)
    line, _, err := r.ReadLine()

    fmt.Println(string(line))

    var bytes []byte
    n, _ := stdout.Read(bytes)
    fmt.Println(n)
    fmt.Println(bytes)

    //stdin.Write([]byte("exit"))
    stdin.Write([]byte(""))

}
