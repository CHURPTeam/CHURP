package commitpbc

import (
	"bytes"
	"encoding/gob"
	"math/rand"
	"testing"
	"time"

	"github.com/bl4ck5un/ChuRP/src/utils/conv"
	"github.com/bl4ck5un/ChuRP/src/utils/polyring"
	"github.com/ncw/gmp"
	"github.com/stretchr/testify/assert"
)

var poly = polyring.FromVec(0, 2, 3, 4, 5, 6)
var poly2 = polyring.FromVec(11, 12, 13, 14, 15, 16)

func TestParams_String(t *testing.T) {
	println(Curve.Params.String())
}

func TestCommit(t *testing.T) {
	comm := NewPolyCommit(poly)
	assert.True(t, comm.Verify(poly))
}

func TestPolyCommit_Verify(t *testing.T) {
	comm := NewPolyCommit(poly)
	assert.True(t, comm.Verify(poly))
}

func TestPolyCommit_Gob(t *testing.T) {
	comm := NewPolyCommit(poly)
	assert.True(t, comm.Verify(poly))

	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(comm)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	dec := gob.NewDecoder(&buf)

	var commNew PolyCommit

	err = dec.Decode(&commNew)
	assert.Nil(t, err)

	assert.True(t, commNew.Equals(comm), "decoding")
}

func TestPolyCommit_VerifyEval(t *testing.T) {
	x := gmp.NewInt(15623523536)
	y := gmp.NewInt(0)
	poly.EvalMod(x, Curve.Ngmp, y)

	comm := NewPolyCommit(poly)
	r := comm.VerifyEval(conv.GmpInt2BigInt(x), conv.GmpInt2BigInt(y))

	assert.True(t, r)
}

func TestAdditiveHomomorphism(t *testing.T) {
	com1 := NewPolyCommit(poly)
	com2 := NewPolyCommit(poly2)

	poly3 := polyring.Polynomial{}
	poly3.Add(poly, poly2)
	poly3.Mod(Curve.Ngmp)

	comm3 := AdditiveHomomorphism(com1, com2)
	assert.True(t, comm3.Verify(poly3))
}

const bigPolyDegree = 100

var rnd = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

func BenchmarkVerifyEval(b *testing.B) {
	poly100, err := polyring.NewRand(bigPolyDegree, rnd, Curve.Ngmp)
	assert.Nil(b, err)

	// x is a random point
	x := new(gmp.Int)
	x.Rand(rnd, Curve.Ngmp)

	y := gmp.NewInt(0)
	poly100.EvalMod(x, Curve.Ngmp, y)

	comm := NewPolyCommit(poly100)

	b.Run("VerifyEval", func(b *testing.B) {
		//start := time.Now()
		for i := 0; i < b.N; i++ {
			comm.VerifyEval(conv.GmpInt2BigInt(x), conv.GmpInt2BigInt(y))
		}
	})
}
