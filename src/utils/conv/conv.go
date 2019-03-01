package conv

import "math/big"
import "github.com/ncw/gmp"

func BigInt2GmpInt(a *big.Int) *gmp.Int {
	b := gmp.NewInt(0)
	b.SetBytes(a.Bytes())

	return b
}

func GmpInt2BigInt(a *gmp.Int) *big.Int {
	b := new(big.Int)
	b.SetBytes(a.Bytes())

	return b
}
