package polypoint

import (
	"github.com/Nik-U/pbc"
	"github.com/ncw/gmp"
)

type PolyPoint struct {
	X       int32
	Y       *gmp.Int
	PolyWit *pbc.Element
}

func NewZeroPoint() *PolyPoint {
	return &PolyPoint{
		X:       0,
		Y:       gmp.NewInt(0),
		PolyWit: nil,
	}
}

func NewPoint(x int32, y *gmp.Int, w *pbc.Element) *PolyPoint {
	return &PolyPoint{
		X:       x,
		Y:       y,
		PolyWit: w,
	}
}
