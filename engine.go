package main


import (
    "os/exec"
    "fmt"
    "regexp"
    "strconv"
    "strings"
)

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

func NewEngine(engineName string) *Engine {

    engineAvailable(engineName)
    //process, _ := launchCmd(engineName, []string{}, []string{})

	return &Engine{
		process:    NewProcessEndpoint(engineName),
		output:     make(chan string),
		input:      make(chan string),
		err:        make(chan bool),
    }
}

type Engine struct {
    process    *ProcessEndpoint
	output     chan string
	input      chan string
	err        chan bool
}

func (eng *Engine) Output() chan string { return eng.output }
func (eng *Engine) Input() chan string { return eng.input }
func (eng *Engine) Err() chan bool { return eng.err }

func (eng *Engine) Terminate() { eng.process.Terminate() }

func (eng *Engine) Start() {
    eng.process.Start()
    go talk(eng.input, eng.output, eng.process.Input(), eng.process.Output())  // add the err channels!
}

func talk(engine_in, engine_out, process_in, process_out chan string) {
    for {
        select {
        case cmd := <-engine_in:
            process_in <- cmd
        case msg := <-process_out:
            engine_out <- msg
        }
    }
}

func ValidSquare(square string) bool {
    // assert valid square
    if regexp.MustCompile(`^[abcdefgh][1-8]$`).MatchString(square) {
        return true
    }
    return false
}

type Fen struct {
    board string
    color string
    castling string
    enpassant string
    halfmoves int
    move int
}

func NewFen(s string) *Fen {
    parts := strings.Split(s, " ")

    board := ""
    color := ""
    castling := ""
    enpassant := ""
    halfmoves := 0
    move := 0

    if len(parts) > 0 { board = parts[0] }
    if len(parts) > 1 { color = parts[1] }
    if len(parts) > 2 { castling = parts[2] }
    if len(parts) > 3 { enpassant = parts[3] }
    if len(parts) > 4 { halfmoves, _ = strconv.Atoi(parts[4]) }
    if len(parts) > 5 { move, _ = strconv.Atoi(parts[5]) }

    return &Fen{board, color, castling, enpassant, halfmoves, move}
}

func (f *Fen) SquareToPiece(square string) string {
    //5k2/4n2p/1q2PBp1/p7/Q4P2/6P1/7P/2rR3K b - -

    rows := strings.Split(f.board, "/")

    // a1 -> (0, 0); h8 -> (7, 7)
    sqCol := square[0] - 97
    sqRow := square[1] - 49

    result := string(ExpandRow(rows[7 - sqRow])[sqCol])
    return result

}

func ExpandRow(r string) string {
    // 5k2 -> xxxxxkxx
    res := ""

    for _, run := range r {
        str := string(run)
        if regexp.MustCompile(`\d`).MatchString(str) {
            num, _ := strconv.Atoi(str)
            res = res + strings.Repeat("x", num)
        } else {
            res = res + str
        }
    }

    return res
}

type Board struct {
    squares map[string]string
}

func NewBoard(fenPosition string) *Board {
    cols := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
    //rows := []string{"1", "2", "3", "4", "5", "6", "7", "8"}
    fenRows := strings.Split(fenPosition, "/") // assert length 8
    sqs := make(map[string]string)

    for i, fenRow := range fenRows {
        expandedFenRow := ExpandRow(fenRow)
        if i > 7 || i < 0 {
            panic("bla")
        }
        row := fmt.Sprintf("%d", 8 - i)
        for j, col := range cols {
            sq := (col + row)
            sqs[sq] = string(expandedFenRow[j])
        }
    }

    return &Board {
        sqs,
    }
}
