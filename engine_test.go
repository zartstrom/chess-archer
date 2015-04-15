package main


import (
    //"fmt"
    "testing"
    "github.com/stretchr/testify/assert"
)


func TestValidSquare_01(t *testing.T) {
    if (ValidSquare("a3") == false) {
        t.Error("ValidSquare did not work as expected.")
    } else {
        t.Log("one test passed.")
    }
}

func TestValidSquare_02(t *testing.T) {
    if (ValidSquare("c9") == true) {
        t.Error("ValidSquare did not work as expected.")
    } else {
        t.Log("one test passed.")
    }
}

func TestValidSquare_03(t *testing.T) {
    if (ValidSquare("h8") == false) {
        t.Error("ValidSquare did not work as expected.")
    } else {
        t.Log("one test passed.")
    }
}

func TestValidSquare_04(t *testing.T) {
    if (ValidSquare("e44") == true) {
        t.Error("ValidSquare did not work as expected.")
    } else {
        t.Log("one test passed.")
    }
}

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
    res := ExpandRow(sample)
    assert.Equal(t, res, "xxxxxkxx")
}

func TestNewBoard_01(t *testing.T) {
    pos := "5k2/4n2p/1q2PBp1/p7/Q4P2/6P1/7P/2rR3K"
    b := NewBoard(pos)
    assert.Equal(t, b.squares["b6"], "q")
    assert.Equal(t, b.squares["a4"], "Q")
    assert.Equal(t, b.squares["h2"], "P")
}
