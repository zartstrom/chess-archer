package main


import (
    //"fmt"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestNewFen_01(t *testing.T) {
    s := "5k2/4n2p/1q2PBp1/p7/Q4P2/6P1/7P/2rR3K b - -"
    f := NewFen(s)

    assert.Equal(t, f.boardString, "5k2/4n2p/1q2PBp1/p7/Q4P2/6P1/7P/2rR3K")
    assert.Equal(t, f.color, "b")
    assert.Equal(t, f.castling, "-")
    assert.Equal(t, f.enpassant, "-")
    assert.Equal(t, f.halfmoves, 0)
    assert.Equal(t, f.move, 0)
}

func TestExpandRow_01(t *testing.T) {
    sample := "5k2"
    res := expandRow(sample)
    assert.Equal(t, res, "xxxxxkxx")
}
