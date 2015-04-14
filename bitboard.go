package main


import (
    "fmt"
    "regexp"
    "strings"
)


func BitMain() {
    fmt.Println("Welcome to the incredible BitBoard challenge!")
    fmt.Println("1. Place a rook on a empty board")
    fmt.Println("And show all possible moves")
    fmt.Println("2. Place black and white obstacles on the board")

    //tests
    //testSquareToBit("d1")
    //testSquareToBit("c3")
    //testSquareToBit("f7")
    //testSquareToBit("f4")
    //testRookMask("e4")
    w_pieces := setPieces("f1", "e2", "b4", "e6")
    b_pieces := setPieces("b7", "c7", "g7", "h7")
    //testBishopMoves("g2", w_pieces, b_pieces)
    //piece := NewPiece("Q", "e7")
    testPieceMoves("B", "b3", w_pieces, b_pieces)
    //pretty(piece.Moves(w_pieces, b_pieces))
    //testKnightMask("h1")
    //testBishopMask("e7")
    bb := NewBitBoard(NewFen(STARTPOSITION))
    bb.Pretty()
}

type BitBoard struct {
    layers map[int]uint64 // layers contain the placements for every piece type
    specials uint8 // special moves castling, en passant go here
}

func NewBitBoard(f *Fen) *BitBoard {
    layers := make(map[int]uint64)

    fenRows := strings.Split(f.boardString, "/") // assert length 8
    // FEN start with 8th rank, flip it
    for i := 7; i >= 0; i-- {
        expandedFenRow := ExpandRow(fenRows[i])

        for j, rn := range expandedFenRow {
            if FEN_TO_PIECE[string(rn)] != 0 {
                shift := uint(8 * (7 - i) + j)
                layers[FEN_TO_PIECE[string(rn)]] |= (1 << shift)
            }
        }
    }
    return &BitBoard{layers, 0}
}

func (bb *BitBoard) GetPiece(square string) int {
    sq_bit := bitSquare(square)
    for piece, layer := range bb.layers {
        if sq_bit & layer != 0 {
            return piece
        }
    }
    return 0 // EMPTY_PIECE
}

func (bb *BitBoard) IsOccupied(square string) bool {
    sq_bit := bitSquare(square)
    return sq_bit & bb.occupiedSquares() != 0
}


// TODO: think about naming occupied vs. occupiedSquares
func (bb *BitBoard) occupied(pieceTypes []int) uint64 {
    result := uint64(0)
    // think about wording piece[s]/pieceType[s]
    for _, pieceType := range pieceTypes {
        result |= bb.layers[pieceType]
    }
    return result
}

// there has to be a better solution
func (bb *BitBoard) makeEmpty(square_bit uint64) {
    for piece, layer := range bb.layers {
        if layer & square_bit != 0 {
            bb.layers[piece] = layer ^ square_bit
        }
    }
}

// TODO: think about naming occupied vs. occupiedSquares
func (bb *BitBoard) occupiedSquares() uint64 {
    return bb.occupied(PIECES)  // PIECES are all (12) piece types
}

func (bb *BitBoard) UpdateBoard(initialSquare, targetSquare string, piece int) {
    init_sq := bitSquare(initialSquare)
    targ_sq := bitSquare(targetSquare)
    // debug
    if init_sq & bb.layers[piece] == 0 {
        panic(fmt.Sprintf("piece not on initial square. %s-%s", initialSquare, targetSquare))
    }

    bb.makeEmpty(init_sq)
    bb.layers[piece] |= targ_sq
}

func (bb *BitBoard) Pretty() {
    var sqs = []string{}

    // func get 64string
    occ := bb.occupied(PIECES)
    for i := 0; i < 64; i++ {
        sqs = append(sqs, " ")
        var sq_bit uint64
        sq_bit = (1 << uint(i))
        if occ & sq_bit == 0 { continue }

        for _, pieceType := range PIECES {
            if bb.layers[pieceType] & sq_bit != 0 {
                sqs[i] = PIECE_TO_FEN[pieceType]
                break
            }
        }
    }
    printBoard(sqs)
}

// just a printer for uint64 for debugging
func pretty(board uint64) {
    var empty  = " "
    var filled = "X"

    var sqs = []string{}
    for i := 0; i < 64; i++ {
        if board & (1 << uint(i)) != 0 {
            sqs = append(sqs, filled)
        } else {
            sqs = append(sqs, empty)
        }
    }
    printBoard(sqs)
}


func printBoard(sqs []string) {
    var sep = "+---+---+---+---+---+---+---+---+\n"
    var result string = ""

    result += sep
    for r := 7; r >= 0; r-- {
        rank := sqs[8 * r: 8 * (r + 1)]
        for _, sym := range rank {
            result += fmt.Sprintf("| %s ", string(sym))
        }
        result += "|\n"
        result += sep
    }
    fmt.Println(result)
}

var SQUARE_REGEX = regexp.MustCompile(`^(?P<fileChar>\w+)(?P<rankChar>\d+)$`)
var fileShift = map[string]int{"a" : 0, "b" : 1, "c" : 2, "d": 3, "e" : 4, "f" : 5, "g" : 6, "h": 7,}
var rankShift = map[string]int{"1" : 0, "2" : 1, "3" : 2, "4": 3, "5" : 4, "6" : 5, "7" : 6, "8": 7,}

func squareToCoords(square string) (int, int) {
    // a1 -> (0, 0), h1 -> (0, 7), h8 -> (7, 7)
    match := SQUARE_REGEX.FindStringSubmatch(square)

    subs := make(map[string]string)
    for i, name := range SQUARE_REGEX.SubexpNames() {
        subs[name] = match[i]
    }
    return rankShift[subs["rankChar"]], fileShift[subs["fileChar"]]
}

func bitSquare(square string) uint64 {
    rankNr, fileNr := squareToCoords(square)
    shift := uint(8 * rankNr + fileNr)
    return (1 << shift)
}

// setPieces takes squares as a parameter and returns a uint64
// with bits turned on for the corresponding squares
func setPieces(squares ...string) uint64 {
    var result uint64 = 0
    for _, square := range squares {
        sq_bit := bitSquare(square)
        result |= sq_bit
    }
    return result
}

// numSquare return the position of a square in 64 integer numbered from 0 to 63
// Example: c1 -> 100 -> 2; a2 -> 1 0000 0000 -> 8
func numSquare(square string) int {
    rankNr, fileNr := squareToCoords(square)
    return 8 * rankNr + fileNr
}

func testSquareToBit(square string) {
    b := bitSquare(square)
    fmt.Printf("\"%s\"\n", square)
    pretty(b)
    fmt.Println()
}

func shift(number uint64, steps int) uint64 {
    if steps > 0 {
        return number << uint(steps)
    } else {
        return number >> uint(-steps)
    }
}


const DELTA_N =  8
const DELTA_E =  1
const DELTA_S = -8
const DELTA_W = -1
var DELTAS_ROOK = []int{DELTA_N, DELTA_E, DELTA_S, DELTA_W}

const DELTA_NW = DELTA_N + DELTA_W
const DELTA_NE = DELTA_N + DELTA_E
const DELTA_SE = DELTA_S + DELTA_E
const DELTA_SW = DELTA_S + DELTA_W
var DELTAS_BISHOP = []int{DELTA_NW, DELTA_NE, DELTA_SE, DELTA_SW}
var DELTAS_QUEEN = append(DELTAS_ROOK, DELTAS_BISHOP...)
var DELTAS_KING = DELTAS_QUEEN

const DELTA_NNW = 2 * DELTA_N + DELTA_W
const DELTA_NNE = 2 * DELTA_N + DELTA_E
const DELTA_EEN = 2 * DELTA_E + DELTA_N
const DELTA_EES = 2 * DELTA_E + DELTA_S
const DELTA_SSE = 2 * DELTA_S + DELTA_E
const DELTA_SSW = 2 * DELTA_S + DELTA_W
const DELTA_WWS = 2 * DELTA_W + DELTA_S
const DELTA_WWN = 2 * DELTA_W + DELTA_N
var DELTAS_KNIGHT = []int{
    DELTA_NNW, DELTA_NNE,
    DELTA_EEN, DELTA_EES,
    DELTA_SSE, DELTA_SSW,
    DELTA_WWS, DELTA_WWN,
}

type Mover interface {
    Deltas() []int
    Moves(w_pieces, b_pieces uint64) uint64
    Mask() uint64
    //SquareStr() string
    //Name() string
}

func NewPiece(ptype string, square_str string) Mover {
    if ptype == "R" || ptype == "r" {
        return NewRook(square_str)
    } else if ptype == "B" || ptype == "b" {
        return NewBishop(square_str)
    } else if ptype == "Q" || ptype == "q" {
        return NewQueen(square_str)
    } else {
        panic("More pieces coming soon...")
    }
}

type Rook struct {
    // probably only square_bit is necessary
    square_str string
    square_bit uint64
    square_num int
}

func NewRook(square_str string) *Rook {
    return &Rook{square_str, bitSquare(square_str), numSquare(square_str)}
}

func (p *Rook) Name() string { return "Rook" }
func (p *Rook) SquareStr() string { return p.square_str }
func (p *Rook) Deltas() []int { return DELTAS_ROOK }
func (p *Rook) Mask() uint64 { return rookMask(p.square_str) }
func (p *Rook) Moves(hero_pieces uint64, opp_pieces uint64) uint64 {
    return getSlidingMoves(p.Deltas(), p.Mask(), p.square_bit, hero_pieces, opp_pieces)
}

type Bishop struct {
    // probably only square_bit is necessary
    square_str string
    square_bit uint64
    square_num int
}

func NewBishop(square_str string) *Bishop {
    return &Bishop{square_str, bitSquare(square_str), numSquare(square_str)}
}

func (p *Bishop) Name() string { return "Bishop" }
func (p *Bishop) SquareStr() string { return p.square_str }
func (p *Bishop) Deltas() []int { return DELTAS_BISHOP }
func (p *Bishop) Mask() uint64 { return bishopMask(p.square_str) }
func (p *Bishop) Moves(hero_pieces uint64, opp_pieces uint64) uint64 {
    return getSlidingMoves(p.Deltas(), p.Mask(), p.square_bit, hero_pieces, opp_pieces)
}

type Queen struct {
    // probably only square_bit is necessary
    square_str string
    square_bit uint64
    square_num int
}

func NewQueen(square_str string) *Queen {
    return &Queen{square_str, bitSquare(square_str), numSquare(square_str)}
}

func (p *Queen) Name() string { return "Queen" }
func (p *Queen) SquareStr() string { return p.square_str }
func (p *Queen) Deltas() []int { return DELTAS_QUEEN }
func (p *Queen) Mask() uint64 { return rookMask(p.square_str) + bishopMask(p.square_str) }
func (p *Queen) Moves(hero_pieces uint64, opp_pieces uint64) uint64 {
    return getSlidingMoves(p.Deltas(), p.Mask(), p.square_bit, hero_pieces, opp_pieces)
}

func getSlidingMoves(deltas []int, mask uint64, square_bit uint64, hero_pieces uint64, opp_pieces uint64) uint64 {
    result := uint64(0)
    for _, delta := range deltas {
        x := square_bit
        for {
            x = shift(x, delta)
            if (x & mask) == 0 {
                // move over the edge of the board
                break

            } else if (x & hero_pieces) != 0 {
                // hit a piece of same color
                break

            } else if (x & opp_pieces) != 0 {
                // hit a opponents piece
                result |= x
                break

            } else {
                // yes, we can move there
                result |= x
            }
        }
    }
    return result
}

func testPieceMoves(ptype string, square_str string, hero_pieces uint64, opp_pieces uint64) {
    p := NewPiece(ptype, square_str)
    b := p.Moves(hero_pieces, opp_pieces)
    fmt.Println("hero pieces:")
    pretty(hero_pieces)
    fmt.Println("opponents pieces:")
    pretty(opp_pieces)
    fmt.Printf("%s on %s can move to:\n", ptype, square_str)
    pretty(b)
}

func rookMask(square string) uint64 {
    rankNr, fileNr := squareToCoords(square)
    fileBB := filesBB[fileNr]
    rankBB := ranksBB[rankNr]
    return fileBB ^ rankBB
}

func testRookMask(square string) {
    fmt.Printf("Rook mask on \"%s\":\n", square)
    pretty(rookMask(square))
}

func bishopMoves(square string, hero_pieces uint64, opp_pieces uint64) uint64 {
    mask := bishopMask(square)
    sq_bit := bitSquare(square)

    return getSlidingMoves(DELTAS_BISHOP, mask, sq_bit, hero_pieces, opp_pieces)
}

func testBishopMoves(square string, hero_pieces uint64, opp_pieces uint64) {
    b := bishopMoves(square, hero_pieces, opp_pieces)
    fmt.Println("my pieces:")
    pretty(hero_pieces)
    fmt.Println("opponents pieces:")
    pretty(opp_pieces)
    fmt.Printf("Bishop on %s can move to:\n", square)
    pretty(b)
}

// bishopMask returns the set of possible squares that a bishop can reach from a given square in a bitboard.
// TODO: think about precomputing every possible mask
func bishopMask(square string) uint64 {
    sq_bit := bitSquare(square)
    sq_num := numSquare(square) // sq_nr in 0..63
    result := uint64(0)

    for _, delta := range DELTAS_BISHOP {
        x_bit := sq_bit
        x_num := sq_num
        for {
            x_bit = shift(x_bit, delta)
            mod_diff := (x_num % 8) - ((x_num + delta) % 8)
            x_num = x_num + delta
            if mod_diff < 0 { mod_diff = -mod_diff }  // abs

            if mod_diff > 1 || x_num < 0 || x_num > 63 {
                break
            } else {
                result |= x_bit
            }
        }
    }
    return result
}

func testBishopMask(square string) {
    fmt.Printf("Bishop mask on \"%s\":\n", square)
    pretty(bishopMask(square))
}

// knightMask returns all the squares a knight can reach from a given square on an empty board.
func knightMask(square string) uint64 {
    sq_num := numSquare(square) // sq_num in 0..63
    result := uint64(0)
    for _, delta := range DELTAS_KNIGHT {
        mod_diff := (sq_num % 8) - ((sq_num + delta) % 8)
        if mod_diff < 0 { mod_diff = -mod_diff }  // abs
        if sq_num + delta >= 0 && sq_num + delta <= 63 && mod_diff < 3 {
            result |= (1 << uint(sq_num + delta))
        }
    }
    return result
}

func testKnightMask(square string) {
    fmt.Printf("Knight mask on \"%s\":\n", square)
    pretty(knightMask(square))
}



const FileABB = 0x0101010101010101
const FileBBB = FileABB << 1;
const FileCBB = FileABB << 2;
const FileDBB = FileABB << 3;
const FileEBB = FileABB << 4;
const FileFBB = FileABB << 5;
const FileGBB = FileABB << 6;
const FileHBB = FileABB << 7;

var filesBB = map[int]uint64{
    0: FileABB,
    1: FileBBB,
    2: FileCBB,
    3: FileDBB,
    4: FileEBB,
    5: FileFBB,
    6: FileGBB,
    7: FileHBB,
}

const Rank1BB = 0xFF;
const Rank2BB = Rank1BB << (8 * 1);
const Rank3BB = Rank1BB << (8 * 2);
const Rank4BB = Rank1BB << (8 * 3);
const Rank5BB = Rank1BB << (8 * 4);
const Rank6BB = Rank1BB << (8 * 5);
const Rank7BB = Rank1BB << (8 * 6);
const Rank8BB = Rank1BB << (8 * 7);

var ranksBB = map[int]uint64{
    0: Rank1BB,
    1: Rank2BB,
    2: Rank3BB,
    3: Rank4BB,
    4: Rank5BB,
    5: Rank6BB,
    6: Rank7BB,
    7: Rank8BB,
}

