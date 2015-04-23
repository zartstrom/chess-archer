package main


import (
    "fmt"
    "regexp"
    "strings"
    "strconv"
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
    //w_pieces := setPieces("f1", "e2", "b4", "e6")
    //b_pieces := setPieces("b7", "c7", "g7", "h7")
    //testBishopMoves("g2", w_pieces, b_pieces)
    //piece := NewPiece("Q", "e7")
    //broken: testPieceMoves("B", "b3", w_pieces, b_pieces)
    //pretty(piece.Moves(w_pieces, b_pieces))
    //testKnightMask("h1")
    //testBishopMask("e7")
    board := NewBitBoard(NewFen(STARTPOSITION))
    board.Pretty()
}

type BitBoard struct {
    layers map[PieceType]uint64 // layers contain the placements for every piece type
    specials uint8 // special moves castling, en passant go here
}

func NewBitBoard(f *Fen) *BitBoard {
    layers := make(map[PieceType]uint64)

    fenRows := strings.Split(f.boardString, "/") // assert length 8
    // FEN start with 8th rank, flip it
    for i := 7; i >= 0; i-- {
        expandedFenRow := expandRow(fenRows[i])

        for j, rn := range expandedFenRow {
            if FEN_TO_PIECE[string(rn)] != 0 {
                shift := uint(8 * (7 - i) + j)
                layers[FEN_TO_PIECE[string(rn)]] |= (1 << shift)
            }
        }
    }
    return &BitBoard{layers, 0}
}

func expandRow(r string) string {
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

func NewBitBoardStart() *BitBoard {
    return NewBitBoard(NewFen(STARTPOSITION))
}

func (board *BitBoard) GetPiece(square *Square) PieceType {
    sq_bit := square.bit()
    for pieceType, layer := range board.layers {
        if sq_bit & layer != 0 {
            return pieceType
        }
    }
    return NO_PIECE
}

func (board *BitBoard) whitePieces() uint64 {
    return board.get(WHITE_PIECES)
}

func (board *BitBoard) blackPieces() uint64 {
    return board.get(BLACK_PIECES)
}

// IsCapture checks for all pieces except pawns if the move is a capture
func (board *BitBoard) IsCapture(pieceType PieceType, targetSquare *Square) bool {
    if pieceType.is(PAWN) { panic("No pawns allowed here!") }

    if board.isOccupied(targetSquare) { return true } else { return false }
}

// IsPawnCapture returns a triple (isCapture bool, isEnPassant bool, enPassantSquare uint64)
func (board *BitBoard) IsPawnCapture(pieceType PieceType, initialSquare, targetSquare *Square) (bool, bool, uint64) {
    if !pieceType.is(PAWN) { panic("Only pawns allowed here!") }

    if board.isOccupied(targetSquare) {
        return true, false, 0
    }

    fromFile, fromRank := initialSquare.coords()
    toFile,   _        := targetSquare.coords()

    if fromFile == toFile {
        // forward move
        return false, false, 0
    } else {
        // en passant
        shift := uint(8 * fromRank + toFile)
        return true, true, (1 << shift)
    }
}

// GetUnambiguity provides information which piece moved to a certain square
// It gives the "f" in "Rfc1" or the "5" in "N5c4"
func (board *BitBoard) GetUnambiguity(pieceType PieceType, initialSquare, targetSquare *Square) string {
    layer := board.layers[pieceType]
    rest := layer ^ initialSquare.bit()
    positions := findBitPositions(rest)
    pieces := []Mover{}

    // TODO: clean it up
    res_file := ""
    res_rank := ""
    for _, num := range positions {
        // TODO: this initialization is pretty awkward
        square := NewSquareByNum(num)
        p := NewPiece(pieceType, square)
        pieces = append(pieces, p)

        // TODO: resolve (white, black) -> (hero, opp) mismatch
        if p.Moves(board.whitePieces(), board.blackPieces()) & targetSquare.bit() != 0 {
            if square.file != initialSquare.file {
                res_file = initialSquare.file
            } else {
                res_rank = initialSquare.rank
            }
        }
    }
    return res_file + res_rank
}

// findBitPositions returns the positions of 1s in a number
// A 8-bit example: 10000110 -> [1, 2, 7]
// TODO: think about where to move this two functions
func findBitPositions(t uint64) []int {
    result := []int{}
    return find(t, 0, 6, result) // 64bit -> depth = 6
}

func find(t uint64, position int, depth int, result []int) []int {
    // position 0 // depth = 6 // shift = 32
    if t == 0 {
        return result
    } else if depth == 0 {
        if t == 1 {
            result = append(result, position)
        }
        return result
    } else {
        shift := 1 << uint(depth - 1)
        ones  := (uint64(1) << uint(shift)) - 1
        left  := t >> uint(shift)
        right := t & ones

        return find(left, position + shift, depth - 1, find(right, position, depth - 1, result))
    }
}

func (board *BitBoard) isOccupied(square *Square) bool {
    return square.bit() & board.occupiedSquares() != 0
}


// TODO: think about naming get vs. occupiedSquares
func (board *BitBoard) get(pieceTypes []PieceType) uint64 {
    result := uint64(0)
    // think about wording piece[s]/pieceType[s]
    for _, pieceType := range pieceTypes {
        result |= board.layers[pieceType]
    }
    return result
}

// there has to be a better solution
func (board *BitBoard) clearSquare(square_bit uint64) {
    for pieceType, layer := range board.layers {
        if layer & square_bit != 0 {
            board.layers[pieceType] = layer ^ square_bit
        }
    }
}

// TODO: think about naming get vs. occupiedSquares
func (board *BitBoard) occupiedSquares() uint64 {
    return board.get(PIECES)  // PIECES are all (12) piece types
}

func (board *BitBoard) UpdateBoard(move *Move) {
    if move.isCastling {
        board.handleCastling(move.castlingType)
        return
    }
    init_sq := move.initialSquare.bit()
    targ_sq := move.targetSquare.bit()
    // assert there is piece on initial square
    if init_sq & board.layers[move.pieceType] == 0 {
        panic(fmt.Sprintf("piece not on initial square. %s-%s", move.initialSquare, move.targetSquare))
    }

    board.clearSquare(targ_sq)
    board.move(move.pieceType, init_sq, targ_sq)
    if move.pieceType.is(PAWN) {
        if move.isEnPassant {
            board.clearSquare(move.enPassantSquare)
        }
        if move.isPromotion {
            board.clearSquare(targ_sq)
            board.layers[move.promotionPiece] |= targ_sq
        }
    }
}

func (board *BitBoard) handleCastling(castlingType CastlingType) {
    if castlingType.is(WHITE_CASTLING_SHORT) {
        board.move(WHITE_KING, E1, G1)
        board.move(WHITE_ROOK, H1, F1)
    } else if castlingType.is(BLACK_CASTLING_SHORT) {
        board.move(BLACK_KING, E8, G8)
        board.move(BLACK_ROOK, H8, F8)
    } else if castlingType.is(WHITE_CASTLING_LONG) {
        board.move(WHITE_KING, E1, C1)
        board.move(WHITE_ROOK, A1, D1)
    } else if castlingType.is(BLACK_CASTLING_LONG) {
        board.move(BLACK_KING, E8, C8)
        board.move(BLACK_ROOK, A8, D8)
    }
}

func (board *BitBoard) move(pieceType PieceType, from uint64, to uint64) {
    board.layers[pieceType] ^= from
    board.layers[pieceType] |= to
}

// TODO: needs some refactoring
func (board *BitBoard) isCheck(color Color) bool {
    var pieceTypes []PieceType
    var kingbit uint64
    if color == WHITE {
        pieceTypes = WHITE_PIECES
        kingbit = board.layers[BLACK_KING]
    } else {
        pieceTypes = BLACK_PIECES
        kingbit = board.layers[WHITE_KING]
    }

    whitePieces := board.whitePieces()
    blackPieces := board.blackPieces()

    attacked := uint64(0)
    for _, pieceType := range pieceTypes {
        if pieceType.is(PAWN) || pieceType.is(KING) {
            continue  // still need to implement moves for king and pawn
        }

        pieces := []Mover{}
        positions := findBitPositions(board.layers[pieceType])
        for _, position := range positions {
            sq := NewSquareByNum(position)
            pieces = append(pieces, NewPiece(pieceType, sq))
        }

        for _, piece := range pieces {
            attacked |= piece.Moves(whitePieces, blackPieces)
        }
    }

    return kingbit & attacked == kingbit
}

func (board *BitBoard) Pretty() {
    var sqs = []string{}

    // func get 64string
    occ := board.get(PIECES)
    for i := 0; i < 64; i++ {
        sqs = append(sqs, " ")
        var sq_bit uint64
        sq_bit = (1 << uint(i))
        if occ & sq_bit == 0 { continue }

        for _, pieceType := range PIECES {
            if board.layers[pieceType] & sq_bit != 0 {
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

// setPieces takes squares as a parameter and returns a uint64
// with bits turned on for the corresponding squares
func setPieces(squares ...string) uint64 {
    var result uint64 = 0
    for _, square := range squares {
        sq := NewSquare(square)
        result |= sq.bit()
    }
    return result
}

func testSquareToBit(square string) {
    b := NewSquare(square).bit()
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
    Moves(white_pieces, black_pieces uint64) uint64
    Mask() uint64
    Square() *Square
    Color() Color
}

func NewPiece(pieceType PieceType, square *Square) Mover {
    //fmt.Println("TTTTTTTTTTTTTTTTTT")
    //fmt.Printf("%d\n", pieceType)
    //fmt.Printf(square.name)
    //fmt.Printf("\n")
    color := pieceType.color()
    if pieceType.is(ROOK) {
        return NewRook(square, color)
    } else if pieceType.is(BISHOP) {
        return NewBishop(square, color)
    } else if pieceType.is(QUEEN) {
        return NewQueen(square, color)
    } else if pieceType.is(KNIGHT) {
        return NewKnight(square, color)
    } else {
        panic("More pieces coming soon...")
    }
}

type Knight struct {
    square *Square
    color Color
}

func NewKnight(square *Square, color Color) *Knight {
    return &Knight{square, color}
}

func (p *Knight) Name() string { return "Knight" }
func (p *Knight) Square() *Square { return p.square }
func (p *Knight) Color() Color { return p.color }
func (p *Knight) Deltas() []int { return DELTAS_KNIGHT }
func (p *Knight) Mask() uint64 { return knightMask(p.square) }
func (p *Knight) Moves(white_pieces uint64, black_pieces uint64) uint64 {
    if p.Color() == WHITE {
        return p.Mask() ^ (p.Mask() & white_pieces)
    } else {
        return p.Mask() ^ (p.Mask() & black_pieces)
    }
}


type Rook struct {
    square *Square
    color Color
}

func NewRook(square *Square, color Color) *Rook {
    return &Rook{square, color}
}

func (p *Rook) Name() string { return "Rook" }
func (p *Rook) Square() *Square { return p.square }
func (p *Rook) Color() Color { return p.color }
func (p *Rook) Deltas() []int { return DELTAS_ROOK }
func (p *Rook) Mask() uint64 { return rookMask(p.square) }
func (p *Rook) Moves(white_pieces uint64, black_pieces uint64) uint64 {
    return getSlidingMoves(p, white_pieces, black_pieces)
}

type Bishop struct {
    square *Square
    color Color
}

func NewBishop(square *Square, color Color) *Bishop {
    return &Bishop{square, color}
}

func (p *Bishop) Name() string { return "Bishop" }
func (p *Bishop) Square() *Square { return p.square }
func (p *Bishop) Color() Color { return p.color }
func (p *Bishop) Deltas() []int { return DELTAS_BISHOP }
func (p *Bishop) Mask() uint64 { return bishopMask(p.square) }
func (p *Bishop) Moves(white_pieces uint64, black_pieces uint64) uint64 {
    return getSlidingMoves(p, white_pieces, black_pieces)
}

type Queen struct {
    square *Square
    color Color
}

func NewQueen(square *Square, color Color) *Queen {
    return &Queen{square, color}
}

func (p *Queen) Name() string { return "Queen" }
func (p *Queen) Square() *Square { return p.square }
func (p *Queen) Color() Color { return p.color }
func (p *Queen) Deltas() []int { return DELTAS_QUEEN }
func (p *Queen) Mask() uint64 { return rookMask(p.square) + bishopMask(p.square) }
func (p *Queen) Moves(white_pieces uint64, black_pieces uint64) uint64 {
    return getSlidingMoves(p, white_pieces, black_pieces)
}

func getSlidingMoves(piece Mover, white_pieces uint64, black_pieces uint64) uint64 {
    deltas := piece.Deltas()
    mask := piece.Mask()
    square_bit := piece.Square().bit()

    hero_pieces := white_pieces
    opp_pieces  := black_pieces

    if piece.Color() == BLACK {
        hero_pieces = black_pieces
        opp_pieces  = white_pieces
    }

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

func testPieceMoves(pieceType PieceType, square *Square, white_pieces uint64, black_pieces uint64) {
    p := NewPiece(pieceType, square)
    b := p.Moves(white_pieces, black_pieces)
    fmt.Println("white pieces:")
    pretty(white_pieces)
    fmt.Println("black pieces:")
    pretty(black_pieces)
    fmt.Printf("%s on %s can move to:\n", pieceType, square)
    pretty(b)
}

func rookMask(square *Square) uint64 {
    fileNr, rankNr := square.coords()
    fileBB := filesBB[fileNr]
    rankBB := ranksBB[rankNr]
    return fileBB ^ rankBB
}

func testRookMask(square_str string) {
    fmt.Printf("Rook mask on \"%s\":\n", square_str)
    square := NewSquare(square_str)
    pretty(rookMask(square))
}

// bishopMask returns the set of possible squares that a bishop can reach from a given square in a bitboard.
// TODO: think about precomputing every possible mask
func bishopMask(square *Square) uint64 {
    sq_bit := square.bit()
    sq_num := square.num() // sq_nr in 0..63
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

func testBishopMask(square_str string) {
    fmt.Printf("Bishop mask on \"%s\":\n", square_str)
    square := NewSquare(square_str)
    pretty(bishopMask(square))
}

// knightMask returns all the squares a knight can reach from a given square on an empty board.
func knightMask(square *Square) uint64 {
    sq_num := square.num() // sq_num in 0..63
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

func testKnightMask(square_str string) {
    fmt.Printf("Knight mask on \"%s\":\n", square_str)
    square := NewSquare(square_str)
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

