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
    go talk(eng.input, eng.output,
            eng.process.Input(), eng.process.Output(),
            eng.err, eng.process.Err())
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

func talk(engine_in, engine_out, process_in, process_out chan string, engine_err, process_err chan bool) {
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
        case <-engine_err:

            return
        case <-process_err:
            return
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
    // learn how to copy struct BitBoard because it is changed here
    // board_local := Voodoo(board)

    moves := []*Move{}
    for _, uciMove := range strings.Split(line, " ") {
        move := NewMove(uciMove, board)
        board.UpdateBoard(move)
        moves = append(moves, move)
    }

    fmt.Println("\n")
    fmt.Println(line)
    board.Pretty()
    return styleLine(moves, moveNumber, whiteToMove, "")
}

func styleLine(moves []*Move, moveNumber int, whiteToMove bool, result string) string {
    if len(moves) == 0 { return "" } // defensive programming

    if result != "" {
        result += " "
    }

    if whiteToMove {
        result += fmt.Sprintf("%d.", moveNumber)
    } else if !whiteToMove && result == "" {
        // first move in line is a black move
        result += fmt.Sprintf("%d...", moveNumber)
    }

    result += styleMove(moves[0])

    if len(moves) == 1 {
        return result
    } else {
        if !whiteToMove { moveNumber += 1 }
        return styleLine(moves[1:], moveNumber, !whiteToMove, result)
    }
}

func styleMove(move *Move) string {
    if move.isCastling {
        return CASTLING_TO_STRING[move.castlingType]
    }

    pieceSymbol     := PIECE_TO_ALGEBRAIC[move.piece]
    captureString   := getCaptureString(move)
    promotionString := getPromotionString(move)

    return pieceSymbol + captureString + move.targetSquare + promotionString
}

func getCaptureString(move *Move) string {
    if !move.isCapture {
        return ""
    } else if typeOf(move.piece, PAWN) {
        return fmt.Sprintf("%sx", string(move.initialSquare[:1]))
    } else {
        return "x"
    }
}

func getPromotionString(move *Move) string {
    if move.isPromotion {
        return PIECE_TO_ALGEBRAIC[move.promotionPiece]
    }
    return ""
}

type Move struct {
    uciMove string
    initialSquare string
    targetSquare string
    piece int

    isCastling bool
    castlingType int

    isCapture bool
    isEnPassant bool
    enPassantSquare uint64

    isPromotion bool
    promotionPiece int
}

func NewMove(uciMove string, board *BitBoard) *Move {
    initialSquare   := string(uciMove[:2])
    targetSquare    := string(uciMove[2:4])
    piece           := board.GetPiece(initialSquare)
    isCastling      := false
    castlingType    := NO_CASTLING
    isCapture       := false
    isEnPassant     := false
    enPassantSquare := uint64(0)
    isPromotion     := false
    promotionPiece  := NO_PIECE

    if typeOf(piece, KING) {
        isCastling, castlingType = checkCastling(uciMove, piece)
    }
    if typeOf(piece, PAWN) {
        isCapture, isEnPassant, enPassantSquare = board.IsPawnCapture(piece, initialSquare, targetSquare)
        isPromotion, promotionPiece = checkPromotion(uciMove, targetSquare)
    } else {
        isCapture = board.IsCapture(piece, targetSquare)
    }

    return &Move{uciMove, initialSquare, targetSquare, piece, isCastling,
        castlingType, isCapture, isEnPassant, enPassantSquare, isPromotion, promotionPiece}
}

// typeOf takes a some piece and checks for piece type
func typeOf(somePiece int, piece int) bool {
    return somePiece % 8 == piece
}

func checkCastling(move string, piece int) (bool, int) {
    if piece != WHITE_KING && piece != BLACK_KING {
        return false, NO_CASTLING
    }
    if move == "e1g1" {
        return true, WHITE_CASTLING_SHORT
    } else if move == "e8g8" {
        return true, BLACK_CASTLING_SHORT
    } else if move == "e1c1" {
        return true, WHITE_CASTLING_LONG
    } else if move == "e8c8" {
        return true, BLACK_CASTLING_LONG
    }
    return false, NO_CASTLING
}

func checkPromotion(move string, targetSquare string) (bool, int) {
    if len(move) != 5 { return false, NO_PIECE }

    pieceStr := string(move[4])
    rankNr, _ := squareToCoords(targetSquare)
    var color bool
    if rankNr == 7 {
        color = WHITE
    } else if rankNr == 0 {
        color = BLACK
    } else { panic("invalid rank for promotion") }

    piece := PROMOTION_TO_PIECE[PromotionKey{color, pieceStr}]
    return true, piece
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
