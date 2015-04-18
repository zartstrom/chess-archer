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

func (as *AnalysisState) WhiteToMove() bool {
    if as.fen.color == "w" { return true } else { return false }
}

func (as *AnalysisState) MoveNumber() int {
    return as.fen.move
}

// FEN start position: rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1
const STARTPOSITION = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

func NewAnalysisState(fenString string) *AnalysisState {

    fen := NewFen(fenString)
    return &AnalysisState{
        fen: fen,
    }
}

func talk(engine_in, engine_out, process_in, process_out chan string) {
    //analysisStarted := false
    as := NewAnalysisState(STARTPOSITION)

    for {
        select {
        case cmd := <-engine_in:
            as.CmdUpdate(cmd)
            process_in <- cmd
        case msg := <-process_out:
            if as.started == false {
                engine_out <- msg
            } else if line := getVariation(msg, as); line != "" {
                engine_out <- line
            }
        }
    }
}

//"info depth 1 seldepth 1 multipv 1 score cp 56 nodes 33 nps 16500 tbhits 0 time 2 pv e2e3 a7a6"
var mainlineRegex = regexp.MustCompile(`.*seldepth (?P<depth>\d+).*score cp (?P<score>\d+).*pv (?P<mainline>.*)$`)

// TODO: AnalysisState as parameter is wrong
func getVariation(msg string, as *AnalysisState) string {
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
    board := NewBitBoard(as.fen)
    prettyLine := PrettyLine(result["mainline"], board, as.MoveNumber(), as.WhiteToMove())
    res := fmt.Sprintf("%3.2f - %s", eval, prettyLine)
    return res
}


// PrettyLine displays a line from the uci engine in a human readable format.
// The function takes a fen position as a parameter to figure out which piece symbol to display
// and keeps track of the position of the pieces down the line.
// - Moves by pieces (here pieces = all pieces except pawns) are denoted by a symbol. 
//   The symbols can be algebraic [R,N,B,Q,K] or figurine [♔ ,♕,♖,♗,♘]
// - Pawn moves don't have a symbol. When a pawn captures, then add the letter of the file
//   where the pawn stood before. (i.e. exd5)
// - Captures are denoted by "x", i.e. Nxb6, or cxd4
// - Moves should be uniquely identified. I.e. with rooks on a1 and f1 write Rae1.
//   Likewise with knights on e5 and e3 write N5c4.
// - Castling has the symbols "0-0" and "0-0-0"
func PrettyLine(line string, board *BitBoard, moveNumber int, whiteToMove bool) string {
    var res string
    moveNr := moveNumber
    // learn how to copy struct BitBoard because it is changed here
    // board_local := Voodoo(board)

    if !whiteToMove {
        res = fmt.Sprintf("%d...", moveNr)
    }
    // apt for recursion
    for i, move := range strings.Split(line, " ") {
        if whiteToMove {
            moveNr = moveNumber + (i + 1) / 2
            res += fmt.Sprintf("%d.", moveNr)
        }
        initialSquare := string(move[:2])
        targetSquare := string(move[2:4])
        piece := board.GetPiece(initialSquare)
        isCapture, enPassant := board.IsCapture(piece, initialSquare, targetSquare)
        _, promotedPiece := isPromotion(move, whiteToMove)
        castlingType := getCastlingType(move, piece)
        res += (styleMove(move, piece, isCapture, castlingType, promotedPiece) + " ")
        board.UpdateBoard(initialSquare, targetSquare, piece, castlingType, enPassant, promotedPiece)
        whiteToMove = !whiteToMove
    }
    res = res[:len(res) - 1]  // cut final withspace, where is the better solution

    fmt.Println("\n")
    fmt.Println(line)
    board.Pretty()
    return res
}

type Move struct {
    uciMove string
    whiteToMove bool
    initialSquare string
    targetSquare string
    isCapture bool
    castlingType int
    isEnPassant bool
    isPromotion bool
    promotionPiece int
}

func styleMove(move string, piece int, isCapture bool, castlingType int, promoPiece int) string {
    initialSquare := string(move[:2])
    targetSquare := string(move[2:4])

    if castlingType != NO_CASTLING {
        return CASTLING_TO_STRING[castlingType]
    }
    var captureString = getCaptureString(isCapture, piece, initialSquare)

    pieceSymbol := PIECE_TO_ALGEBRAIC[piece]
    promoString := ""
    if promoPiece != 0 {
        promoString = PIECE_TO_ALGEBRAIC[promoPiece]
    }
    return pieceSymbol + captureString + targetSquare + promoString
}

func isPromotion(move string, whiteToMove bool) (bool, int) {
    if len(move) != 5 { return false, NO_PIECE }

    pieceStr := string(move[4])
    piece := PROMOTION_TO_PIECE[PromotionKey{whiteToMove, pieceStr}]
    return true, piece
}

func getCaptureString(isCapture bool, piece int, initialSquare string) string {
    if !isCapture {
        return ""
    } else if piece == WHITE_PAWN || piece == BLACK_PAWN {
        return fmt.Sprintf("%sx", string(initialSquare[:1]))
    } else {
        return "x"
    }
}

func getCastlingType(move string, piece int) int {
    if piece != WHITE_KING && piece != BLACK_KING {
        return NO_CASTLING
    }
    if move == "e1g1" {
        return WHITE_CASTLING_SHORT
    } else if move == "e8g8" {
        return BLACK_CASTLING_SHORT
    } else if move == "e1c1" {
        return WHITE_CASTLING_LONG
    } else if move == "e8c8" {
        return BLACK_CASTLING_LONG
    }
    return NO_CASTLING
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
    color := "w"
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
