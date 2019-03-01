package p521

import (
	"crypto/elliptic"
	"math/big"
)

// Marshal encodes a ECC Point into it's compressed representation
func Marshal(curve elliptic.Curve, x, y *big.Int) []byte {
	byteLen := (curve.Params().BitSize + 7) >> 3

	ret := make([]byte, 1+byteLen)

	// handle point at infinity specially
	if x.BitLen() == 0 && y.BitLen() == 0 {
		ret[0] = 0xff
		return ret
	}

	ret[0] = 2 + byte(y.Bit(0))

	xBytes := x.Bytes()
	copy(ret[1+byteLen-len(xBytes):], xBytes)

	return ret
}

// Unmarshal decodes an ECC Point from any representation
func Unmarshal(curve elliptic.Curve, data []byte) (x, y *big.Int) {
	// handle infinity points specially
	if data[0] == 0xff {
		x = big.NewInt(0)
		y = big.NewInt(0)
		return
	}

	// Split the sign byte from the rest
	sign_byte := uint(data[0])
	x_bytes := data[1:]

	// Convert to big Int.
	x = new(big.Int).SetBytes(x_bytes)

	// We use 3 a couple of times
	three := big.NewInt(3)

	// The params for P256
	c := curve.Params()

	// The equation is y^2 = x^3 - 3x + b
	// x^3, mod P
	x_cubed := new(big.Int).Exp(x, three, c.P)

	// 3x, mod P
	three_X := new(big.Int).Mul(x, three)
	three_X.Mod(three_X, c.P)

	// x^3 - 3x
	y_squared := new(big.Int).Sub(x_cubed, three_X)

	// ... + b mod P
	y_squared.Add(y_squared, c.B)
	y_squared.Mod(y_squared, c.P)

	// Now we need to find the square root mod P.
	// This is where Go's big int library redeems itself.
	y = big.NewInt(0).ModSqrt(y_squared, c.P)
	if y == nil {
		// If this happens then you're dealing with an invalid point.
		// Panic, return an error, whatever you want.
		panic("Invalid point")
		return
	}

	// Finally, check if you have the correct root. If not you want
	// -y mod P
	if y.Bit(0) != sign_byte&1 {
		y.Neg(y)
		y.Mod(y, c.P)
	}

	return x, y
}
