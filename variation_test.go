package main


import (
    //"fmt"
    "testing"
    "github.com/stretchr/testify/assert"
)


func TestPrettyLine_01(t *testing.T) {
    // test castling short
    board := NewBitBoardStart()
    uci_line := "e2e4 e7e5 g1f3 b8c6 f1c4 f8c5 e1g1 g8f6 f1e1 e8g8 g1h1 f8e8 d1e2 g8h8"
    res := PrettyLine(uci_line, board, 1, true)
    assert.Equal(t, "1.e4 e5 2.Nf3 Nc6 3.Bc4 Bc5 4.0-0 Nf6 5.Re1 0-0 6.Kh1 Re8 7.Qe2 Kh8", res)
}

func TestPrettyLine_02(t *testing.T) {
    // test capture and move
    board := NewBitBoardStart()
    uci_line := "e2e4 e7e5 g1f3 g8f6 f3e5 d7d6 e5f3 f6e4"
    res := PrettyLine(uci_line, board, 1, true)
    assert.Equal(t, "1.e4 e5 2.Nf3 Nf6 3.Nxe5 d6 4.Nf3 Nxe4", res)
}

func TestPrettyLine_03(t *testing.T) {
    // test en passant
    board := NewBitBoardStart()
    uci_line := "e2e4 g8f6 e4e5 d7d5 e5d6 f6d5"
    res := PrettyLine(uci_line, board, 1, true)
    assert.Equal(t, "1.e4 Nf6 2.e5 d5 3.exd6 Nd5", res)
}

func TestPrettyLine_04(t *testing.T) {
    // test promotion
    fenString := "6r1/pkp2P2/1p1b4/6n1/8/1P6/PKP5/8 w - -"
    fen := NewFen(fenString)
    board := NewBitBoard(fen)
    uci_line := "f7f8q g5f7 f8g8 a7a5 g8f7"
    res := PrettyLine(uci_line, board, 1, true)
    assert.Equal(t, "1.f8Q Nf7 2.Qxg8 a5 3.Qxf7", res)
}

func TestPrettyLine_05(t *testing.T) {
    // test single move
    board := NewBitBoardStart()
    uci_line := "e2e4"
    res := PrettyLine(uci_line, board, 1, true)
    assert.Equal(t, "1.e4", res)
}

func TestPrettyLine_06(t *testing.T) {
    // test castling long
    board := NewBitBoardStart()
    uci_line := "e2e4 e7e5 g1f3 g8f6 f3e5 d7d6 e5f3 f6e4 d1e2 d8e7 d2d3 e4f6 c1d2 b8c6 b1c3 c8e6 e1c1 e8c8 c1b1 c8b8 d3d4 h7h6 a2a3 a7a6 d1e1 d6d5 d2f4"
    res := PrettyLine(uci_line, board, 1, true)
    assert.Equal(t, "1.e4 e5 2.Nf3 Nf6 3.Nxe5 d6 4.Nf3 Nxe4 5.Qe2 Qe7 6.d3 Nf6 7.Bd2 Nc6 8.Nc3 Be6 9.0-0-0 0-0-0 10.Kb1 Kb8 11.d4 h6 12.a3 a6 13.Re1 d5 14.Bf4", res)
}

func TestPrettyLine_unique_01(t *testing.T) {
    // white rooks on a1 and f1, black knights on c5 and g5
    fenString := "6k1/8/8/2n3n1/8/8/8/R4RK1 w - -"
    fen := NewFen(fenString)
    board := NewBitBoard(fen)
    uci_line := "a1e1 g5e4 e1a1 e4d6 f1c1 c5b7"
    res := PrettyLine(uci_line, board, 1, true)
    assert.Equal(t, "1.Rae1 Nge4 2.Ra1 Nd6 3.Rfc1 Ncb7", res)
}

func TestPrettyLine_unique_02(t *testing.T) {
    // white queens on a4, a6, a8, e4, e6, e8 and black knights on g2 and g6, black rooks on h2, h6
    fenString := "Q3Q3/8/Q3Q1nr/8/Q3Q3/8/6nr/1K4k1 w - -"
    fen := NewFen(fenString)
    board := NewBitBoard(fen)
    uci_line := "a6c6 h2h4 c6a6 h4h2 e8c6 g6f4 c6e8 f4g6"
    res := PrettyLine(uci_line, board, 1, true)
    assert.Equal(t, "1.Qa6c6 R2h4 2.Qca6 Rh2 3.Qe8c6 N6f4 4.Qce8 Ng6", res)
}

func TestPrettyLine_mate_01(t *testing.T) {
    fenString := "8/7p/p3RNk1/4R3/7P/3q2P1/5PK1/8 w - -"
    fen := NewFen(fenString)
    board := NewBitBoard(fen)
    uci_line := "f6h5 g6f7 e6f6 f7g8 e5e8"
    res := PrettyLine(uci_line, board, 1, true)
    assert.Equal(t, "1.Nh5+ Kf7 2.Rf6+ Kg8 3.Re8+", res)  // TODO: need to implement mate
}

func TestPrintMainline_01(t *testing.T) {
    fenString := "8/7p/p3RNk1/4R3/7P/3q2P1/5PK1/8 w - - 0 1"
    uci := "info depth 11 seldepth 8 multipv 1 score mate 3 nodes 2530 nps 253000 tbhits 0 time 10 pv f6h5 g6f7 e6f6 f7g8 e5e8"
    as := NewAnalysisState(fenString)
    res := printMainline(uci, as)
    assert.Equal(t, "#3 - 1.Nh5+ Kf7 2.Rf6+ Kg8 3.Re8+", res)  // TODO: need to implement mate
}
