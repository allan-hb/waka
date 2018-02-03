package gomoku

import . "gopkg.in/ahmetb/go-linq.v3"

var (
	patterns = []struct {
		OffsetX []int32
		OffsetY []int32
	}{
		{[]int32{-2, -1, 0, 1, 2}, []int32{0, 0, 0, 0, 0}},
		{[]int32{0, 0, 0, 0, 0}, []int32{-2, -1, 0, 1, 2}},
		{[]int32{-2, -1, 0, 1, 2}, []int32{-2, -1, 0, 1, 2}},
		{[]int32{-2, -1, 0, 1, 2}, []int32{2, 1, 0, -1, -2}},
	}
)

func (board Board) Judge() (victory string, finally bool) {

	lookup := func(board Board, x, y int32) PieceType {
		for _, pattern := range patterns {
			var pieces []PieceType
			for offset := 0; offset < 5; offset++ {
				posX := x + pattern.OffsetX[offset]
				posY := y + pattern.OffsetY[offset]
				pieces = append(pieces, board.Lookup(posX, posY))
			}

			if From(pieces).AllT(func(piece PieceType) bool { return piece == PieceType_Black }) {
				return PieceType_Black
			}
			if From(pieces).AllT(func(piece PieceType) bool { return piece == PieceType_White }) {
				return PieceType_White
			}
		}
		return PieceType_None
	}

	for y := int32(1); y <= 15; y++ {
		for x := int32(1); x <= 15; x++ {
			val := lookup(board, x, y)
			if val == PieceType_Black {
				return "black", true
			} else if val == PieceType_White {
				return "white", true
			}
		}
	}

	return "none", false
}
