package main


import (
    //"fmt"
    "testing"
    "github.com/stretchr/testify/assert"
)


func TestVariation_01(t *testing.T) {
    board := NewBitBoardStart()
    uci_line := "e2e4 e7e5 g1f3 b8c6 f1c4 f8c5 e1g1 g8f6 f1e1 e8g8 g1h1 f8e8 d1e2 g8h8"
    res := PrettyLine(uci_line, board, 1, true)
    assert.Equal(t, "1.e4 e5 2.Nf3 Nc6 3.Bc4 Bc5 4.0-0 Nf6 5.Re1 0-0 6.Kh1 Re8 7.Qe2 Kh8", res)
    //assert.Equal(t, "1.e4 e5 2.Nf3 Nc6 3.Bc4 Bc5 4.0-0 Nf6 5.Te1", res)
}
