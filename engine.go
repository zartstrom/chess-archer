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

var REGEX_POSITION_FEN = regexp.MustCompile(`position fen (?P<fenString>.*)`)

func (as *AnalysisState) CmdUpdate(cmd string) {
    if regexp.MustCompile(`^go`).MatchString(cmd) {
        as.started = true
    }
    if REGEX_POSITION_FEN.MatchString(cmd) {
        match := REGEX_POSITION_FEN.FindStringSubmatch(cmd)
        as.fen = NewFen(match[1])
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
            } else if line := printMainline(msg, as); line != "" {
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
var mainlineRegex = regexp.MustCompile(`.*seldepth (?P<depth>\d+).*score (?P<score>\w+ \d+).*pv (?P<mainline>.*)$`)

// TODO: AnalysisState as parameter seems wrong
func printMainline(msg string, as *AnalysisState) string {
    if !regexp.MustCompile(`seldepth`).MatchString(msg) {
        return ""
    }
    match := mainlineRegex.FindStringSubmatch(msg)
    result := make(map[string]string)
    for i, name := range mainlineRegex.SubexpNames() {
        result[name] = match[i]
    }
    eval := getEval(result["score"])

    board := NewBitBoard(as.fen)
    prettyLine := PrettyLine(result["mainline"], board, as.MoveNumber(), as.WhiteToMove())
    res := fmt.Sprintf("%s - %s", eval, prettyLine)
    return res
}


var scoreRegex = regexp.MustCompile(`(?P<scoreType>\w+) (?P<val>\d+)`)
func getEval(score string) string {
    match := scoreRegex.FindStringSubmatch(score)
    if len(match) < 2 { return "error in getEval" }

    result := make(map[string]string)
    for i, name := range scoreRegex.SubexpNames() {
        result[name] = match[i]
    }
    if result["scoreType"] == "mate" {
        return fmt.Sprintf("#%s", result["val"])
    } else if result["scoreType"] == "cp" {
        eval, _ := strconv.ParseFloat(result["val"], 64)
        eval = eval / 100
        return fmt.Sprintf("%3.2f", result["val"])
    } else {
        panic(fmt.Sprintf("Unknown score type %s", result["scoreType"]))
    }
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

    color := Color(whiteToMove)
    moves := []*Move{}
    for _, uciMove := range strings.Split(line, " ") {
        move := NewMove(uciMove, board)
        board.UpdateBoard(move)
        move.isCheck = board.isCheck(color)
        moves = append(moves, move)
        color = !color
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
        return move.castlingType.String()
    }

    pieceSymbol     := PIECE_TO_ALGEBRAIC[move.pieceType]

    return pieceSymbol +
        getCaptureString(move) +
        move.unambiguity +
        move.targetSquare.name +
        getPromotionString(move) +
        getCheckString(move)
}

func getCaptureString(move *Move) string {
    if !move.isCapture {
        return ""
    } else if move.pieceType.is(PAWN) {
        return fmt.Sprintf("%sx", string(move.initialSquare.file))
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

func getCheckString(move *Move) string {
    if move.isCheck { return "+" } else { return "" }
}

type Square struct {
    name string
    file string
    rank string
}

var SQUARE_REGEX = regexp.MustCompile(`^(?P<file>\w+)(?P<rank>\d+)$`)

func NewSquare(name string) *Square {
    match := SQUARE_REGEX.FindStringSubmatch(name)

    subexps := make(map[string]string)
    for i, name := range SQUARE_REGEX.SubexpNames() { subexps[name] = match[i] }

    file := subexps["file"]
    rank := subexps["rank"]

    return &Square{name, file, rank}
}

func NewSquareByNum(num int) *Square {
    // TODO: unittest this
    file_int := num % 8
    rank_int := num / 8
    square_str := fmt.Sprintf("%s%d", string(97 + file_int), rank_int + 1)
    return NewSquare(square_str)
}

func (s *Square) bit() uint64 {
    fileNr, rankNr := s.coords()
    shift := uint(8 * rankNr + fileNr)
    return (1 << shift)
}

var FILE_INT = map[string]int{"a" : 0, "b" : 1, "c" : 2, "d": 3, "e" : 4, "f" : 5, "g" : 6, "h": 7,}
var RANK_INT = map[string]int{"1" : 0, "2" : 1, "3" : 2, "4": 3, "5" : 4, "6" : 5, "7" : 6, "8": 7,}

func (s *Square) coords() (int, int) {
    // a1 -> (0, 0), c2 -> (1, 2), h1 -> (0, 7), h8 -> (7, 7)
    return FILE_INT[s.file], RANK_INT[s.rank]
}

// num returns the position of a square in 64 integer numbered from 0 to 63
// Example: c1 -> 100 -> 2; a2 -> 1 0000 0000 -> 8
func (s *Square) num() int {
    file_int, rank_int := s.coords()
    return 8 * rank_int + file_int
}

type Move struct {
    uciMove string
    initialSquare *Square
    targetSquare *Square
    pieceType PieceType

    isCastling bool
    castlingType CastlingType

    isCapture bool
    isEnPassant bool
    enPassantSquare uint64

    isPromotion bool
    promotionPiece PieceType

    isCheck bool

    unambiguity string
}

func NewMove(uciMove string, board *BitBoard) *Move {
    initialSquare   := NewSquare(uciMove[:2])
    targetSquare    := NewSquare(uciMove[2:4])
    pieceType       := board.GetPiece(initialSquare)
    isCastling      := false
    castlingType    := NO_CASTLING
    isCapture       := false
    isEnPassant     := false
    enPassantSquare := uint64(0)
    isPromotion     := false
    promotionPiece  := NO_PIECE
    isCheck         := false
    unambiguity     := "" // is it Nd2 or Nbd2

    if pieceType.is(PAWN) {
        isCapture, isEnPassant, enPassantSquare = board.IsPawnCapture(pieceType, initialSquare, targetSquare)
        isPromotion, promotionPiece = checkPromotion(uciMove, targetSquare)
    } else if pieceType.is(KING) {
        isCastling, castlingType = checkCastling(uciMove, pieceType)
        isCapture = board.IsCapture(pieceType, targetSquare)
    } else {
        isCapture = board.IsCapture(pieceType, targetSquare)
        unambiguity = board.GetUnambiguity(pieceType, initialSquare, targetSquare)
    }

    return &Move{uciMove, initialSquare, targetSquare, pieceType, isCastling, castlingType,
        isCapture, isEnPassant, enPassantSquare, isPromotion, promotionPiece, isCheck, unambiguity}
}

func checkCastling(move string, pieceType PieceType) (bool, CastlingType) {
    if !pieceType.is(KING) {
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

func checkPromotion(move string, targetSquare *Square) (bool, PieceType) {
    if len(move) != 5 { return false, NO_PIECE }

    pieceStr := string(move[4])
    _, rankNr := targetSquare.coords()
    var color Color
    if rankNr == 7 {
        color = WHITE
    } else if rankNr == 0 {
        color = BLACK
    } else { panic("invalid rank for promotion") }

    pieceType := PROMOTION_TO_PIECE[PromotionKey{color, pieceStr}]
    return true, pieceType
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
