package p521

import (
	"bytes"
	"encoding/gob"
	"math/big"
	"testing"

	"../../conv"
	"../../polyring"
	"github.com/ncw/gmp"
	"github.com/stretchr/testify/assert"
)

var poly = polyring.FromVec(0, 2, 3, 4, 5, 6)
var poly2 = polyring.FromVec(11, 12, 13, 14, 15, 16)

func TestECPoint(t *testing.T) {
	x, y := Curve.ScalarBaseMult(big.NewInt(23487263784).Bytes())
	ecp := ECPoint{
		x: x,
		y: y,
	}

	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(ecp)
	assert.Nil(t, err)

	dec := gob.NewDecoder(&buf)
	ecpNew := ECPoint{}

	err = dec.Decode(&ecpNew)
	assert.Nil(t, err)

	println(ecp.String())
	println(ecpNew.String())

	assert.True(t, ecpNew.Equals(ecp))
}

func TestECPointZero(t *testing.T) {
	x, y := Curve.ScalarBaseMult(big.NewInt(0).Bytes())
	ecp := ECPoint{
		x: x,
		y: y,
	}

	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(ecp)
	assert.Nil(t, err)

	dec := gob.NewDecoder(&buf)
	ecpNew := ECPoint{}

	err = dec.Decode(&ecpNew)
	assert.Nil(t, err)

	println(ecp.String())
	println(ecpNew.String())

	assert.True(t, ecpNew.Equals(ecp))
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

	commNew := PolyCommit{}

	err = dec.Decode(&commNew)
	assert.Nil(t, err)

	assert.True(t, commNew.Equals(comm), "decoding")
}

func TestPolyCommit_VerifyEval(t *testing.T) {
	x := gmp.NewInt(15623523536)
	y := gmp.NewInt(0)
	poly.EvalMod(x, conv.BigInt2GmpInt(Curve.Params().N), y)

	comm := NewPolyCommit(poly)
	r := comm.VerifyEval(conv.GmpInt2BigInt(x), conv.GmpInt2BigInt(y))

	assert.True(t, r)
}

func TestAdditiveHomomorphism(t *testing.T) {
	com1 := NewPolyCommit(poly)
	com2 := NewPolyCommit(poly2)

	poly3 := polyring.Polynomial{}
	poly3.Add(poly, poly2)
	poly3.Mod(conv.BigInt2GmpInt(Curve.Params().N))

	comm3 := AdditiveHomomorphism(com1, com2)
	assert.True(t, comm3.Verify(poly3))
}
