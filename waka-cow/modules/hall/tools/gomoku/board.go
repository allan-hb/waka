package gomoku

type PieceType int32

const (
	PieceType_None  PieceType = 0
	PieceType_Black PieceType = 1
	PieceType_White PieceType = 2
)

type Board []PieceType

func (board Board) Play(x, y int32, piece PieceType) {
	pos := (y-1)*15 + x - 1
	if pos < 0 || pos > 224 {
		return
	}
	board[pos] = piece
}

func (board Board) Lookup(x, y int32) PieceType {
	pos := (y-1)*15 + x - 1
	if pos < 0 || pos > 224 {
		return PieceType_None
	}
	return board[pos]
}

func (board Board) ToSlice() []int32 {
	r := make([]int32, 255)
	for i := range board {
		r[i] = int32(board[i])
	}
	return r
}

func NewBoard() Board {
	return make(Board, 225)
}
