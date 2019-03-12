package interpolation

import (
	"errors"

	. "github.com/bl4ck5un/ChuRP/src/utils/polyring"
	"github.com/ncw/gmp"
)

func deduplicate(s []int) []int {
	seen := make(map[int]struct{}, len(s))
	j := 0
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		s[j] = v
		j++
	}
	return s[:j]
}

// LagrangeInterpolate returns a polynomial of specified degree that pass through all points in x and y
func LagrangeInterpolate(degree int, x []*gmp.Int, y []*gmp.Int, mod *gmp.Int) (Polynomial, error) {
	// initialize variables
	tmp, err := New(1)
	if err != nil {
		return Polynomial{}, err
	}

	inter, err := New(degree)
	if err != nil {
		return Polynomial{}, err
	}

	product := NewOne()

	resultPoly, err := New(degree)
	if err != nil {
		return Polynomial{}, err
	}

	denominator := gmp.NewInt(0)

	// tmp(x) = x - x[i]
	tmp.SetCoefficient(1, 1)
	// note only the first degree points are used
	for i := 0; i <= degree; i++ {
		tmp.GetPtrToConstant().Neg(x[i])
		product.MulSelf(tmp)
	}

	for i := 0; i <= degree; i++ {
		denominator.Set(gmp.NewInt(1))
		// compute denominator and numerator

		// tmp = x - x[i]
		tmp.SetCoefficient(1, 1) // i don't think this needed...
		tmp.GetPtrToConstant().Neg(x[i])

		// inner(x) = (x-1)(x_2)...(x-n) except for (x-i)
		err = inter.Div2(product, tmp)
		if err != nil {
			return Polynomial{}, err
		}

		// lambda_i(x) = inner(x) * y[i] / inner(x[i])

		inter.Mod(mod)
		inter.EvalMod(x[i], mod, denominator)

		// panic if denominator == 0
		if 0 == denominator.CmpInt32(0) {
			return Polynomial{}, errors.New("internal error: check duplication in x[]")
		}

		denominator.ModInverse(denominator, mod)
		denominator.Mul(denominator, y[i])
		resultPoly.AddMul(inter, denominator)
	}

	resultPoly.Mod(mod)

	return resultPoly, nil
}
