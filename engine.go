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

type AnalysisState struct {
    started bool
    fen *Fen
}

func (as *AnalysisState) CmdUpdate(cmd string) {
    if regexp.MustCompile(`^go`).MatchString(cmd) {
        as.started = true
    }
}

// FEN start position: rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1
const STARTPOSITION = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

func NewAnalysisState() *AnalysisState {
    return &AnalysisState{
        fen: NewFen(STARTPOSITION),
    }
}

func talk(engine_in, engine_out, process_in, process_out chan string) {
    //analysisStarted := false
    as := NewAnalysisState()

    for {
        select {
        case cmd := <-engine_in:
            as.CmdUpdate(cmd)
            process_in <- cmd
        case msg := <-process_out:
            if as.started == false {
                engine_out <- msg
            } else if line := getVariation(msg, as.fen); line != "" {
                engine_out <- line
            }
        }
    }
}

//"info depth 1 seldepth 1 multipv 1 score cp 56 nodes 33 nps 16500 tbhits 0 time 2 pv e2e3 a7a6"
var mainlineRegex = regexp.MustCompile(`.*seldepth (?P<depth>\d+).*score cp (?P<score>\d+).*pv (?P<mainline>.*)$`)

func getVariation(msg string, fen *Fen) string {
    if !regexp.MustCompile(`seldepth`).MatchString(msg) {
        return ""
    }
    match := mainlineRegex.FindStringSubmatch(msg)
    result := make(map[string]string)
    for i, name := range mainlineRegex.SubexpNames() {
        result[name] = match[i]
    }
    eval, _ := strconv.ParseFloat(result["score"], 64)
    eval = eval / 100
    prettyLine := prettyPrint(result["mainline"], fen) // do it elsewhere
    res := fmt.Sprintf("%3.2f - %s", eval, prettyLine)
    return res
}

func prettyPrint(line string, fen *Fen) string {
    // pretty print a line (=variation)
    var moveNumber int
    var whiteToMove = true
    var res string

    board := NewBoard(fen.boardString)

    if fen.color == "b" {
        res = fmt.Sprintf("%d...", moveNumber)
        whiteToMove = false
    }
    for i, move := range strings.Split(line, " ") {
        if whiteToMove {
            moveNumber = fen.move + (i + 1) / 2
            res += fmt.Sprintf("%d.", moveNumber)
        }
        res += (styleMove(move, board) + " ")
        whiteToMove = !whiteToMove
    }
    return res
}

func isCapture(targetSquare string, board Board) bool {
    return board.squares[targetSquare] != "x"
}

func isCastling(move string) (string, bool) {
    if move == "e1g1" {
        return "0-0", true
    } else if move == "e8g8" {
        return "0-0", true
    } else if move == "e1c8" {
        return "0-0-0", true
    } else if move == "e8c8" {
        return "0-0-0", true
    }
    return "", false


}

func styleMove(move string, board *Board) string {
    initialSquare := string(move[:2])
    targetSquare := string(move[2:4])

    piece := FEN_TO_PIECE[board.squares[initialSquare]]
    if piece == WHITE_KING || piece == BLACK_KING {
        if castlingString, yes := isCastling(move); yes {
            return castlingString
        }
    }
    htmlPiece := PIECE_TO_UTF8[piece]
    return htmlPiece + targetSquare
}

func ValidSquare(square string) bool {
    // assert valid square
    if regexp.MustCompile(`^[abcdefgh][1-8]$`).MatchString(square) {
        return true
    }
    return false
}

type Fen struct {
    boardString string
    color string
    castling string
    enpassant string
    halfmoves int
    move int
}

func NewFen(s string) *Fen {
    parts := strings.Split(s, " ")

    boardString := ""
    color := ""
    castling := ""
    enpassant := ""
    halfmoves := 0
    move := 0

    if len(parts) > 0 { boardString = parts[0] }
    if len(parts) > 1 { color = parts[1] }
    if len(parts) > 2 { castling = parts[2] }
    if len(parts) > 3 { enpassant = parts[3] }
    if len(parts) > 4 { halfmoves, _ = strconv.Atoi(parts[4]) }
    if len(parts) > 5 { move, _ = strconv.Atoi(parts[5]) }

    return &Fen{boardString, color, castling, enpassant, halfmoves, move}
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
