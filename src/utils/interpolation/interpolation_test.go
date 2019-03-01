package interpolation

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	. "../polyring"
	"github.com/ncw/gmp"
	"github.com/stretchr/testify/assert"
)

const POLY_ORDER = 500
const RAND_SEED = 2

var large_str string

func gen_prime(p *gmp.Int, bitnum int) {
	var buffer bytes.Buffer
	for i := 0; i < bitnum; i++ {
		buffer.WriteString("0")
	}

	large_str = "1"
	large_str += buffer.String()

	p.SetString(large_str, 10)
	// No next_prime method in go yet. Placeholder for now
	p.Set(gmp.NewInt(15486511))
	// p.Set(gmp.NewInt(7))
}

func TestLagrangeInterpolate(t *testing.T) {
	p := gmp.NewInt(0)
	gen_prime(p, 256)
	r := rand.New(rand.NewSource(RAND_SEED))

	fmt.Printf("Prime p = %s\n", p.String())

	originalPoly, err := NewRand(POLY_ORDER, r, p)
	assert.Nil(t, err, "New")

	// Test EvalArray
	x := make([]*gmp.Int, POLY_ORDER+1)
	y := make([]*gmp.Int, POLY_ORDER+1)
	VecInit(x)
	VecInit(y)
	VecRand(x, p, r)

	originalPoly.EvalModArray(x, p, y)

	fmt.Println("Finished eval")
	fmt.Println("Starting interpolation")

	reconstructedPoly, err := LagrangeInterpolate(POLY_ORDER, x, y, p)
	assert.Nil(t, err, "New")

	//fmt.Printf("Original Poly ")
	//originalPoly.Print()

	//fmt.Printf("Reconstructed Poly ")
	//reconstructedPoly.Print()
	assert.True(t, reconstructedPoly.IsSame(originalPoly))
}
