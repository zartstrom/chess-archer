package main


import (
    //"fmt"
    "testing"
    "github.com/stretchr/testify/assert"
)


func TestVariation_01(t *testing.T) {
    // test castling short
    board := NewBitBoardStart()
    uci_line := "e2e4 e7e5 g1f3 b8c6 f1c4 f8c5 e1g1 g8f6 f1e1 e8g8 g1h1 f8e8 d1e2 g8h8"
    res := PrettyLine(uci_line, board, 1, true)
    assert.Equal(t, "1.e4 e5 2.Nf3 Nc6 3.Bc4 Bc5 4.0-0 Nf6 5.Re1 0-0 6.Kh1 Re8 7.Qe2 Kh8", res)
}

func TestVariation_02(t *testing.T) {
    // test capture and move
    board := NewBitBoardStart()
    uci_line := "e2e4 e7e5 g1f3 g8f6 f3e5 d7d6 e5f3 f6e4"
    res := PrettyLine(uci_line, board, 1, true)
    assert.Equal(t, "1.e4 e5 2.Nf3 Nf6 3.Nxe5 d6 4.Nf3 Nxe4", res)
}

func TestVariation_03(t *testing.T) {
    // test en passant
    board := NewBitBoardStart()
    uci_line := "e2e4 g8f6 e4e5 d7d5 e5d6 f6d5"
    res := PrettyLine(uci_line, board, 1, true)
    assert.Equal(t, "1.e4 Nf6 2.e5 d5 3.exd6 Nd5", res)
}

func TestVariation_04(t *testing.T) {
    // test promotion
    fenString := "6r1/pkp2P2/1p1b4/6n1/8/1P6/PKP5/8 w - -"
    fen := NewFen(fenString)
    board := NewBitBoard(fen)
    uci_line := "f7f8q g5f7 f8g8 a7a5 g8f7"
    res := PrettyLine(uci_line, board, 1, true)
    assert.Equal(t, "1.f8Q Nf7 2.Qxg8 a5 3.Qxf7", res)
}
