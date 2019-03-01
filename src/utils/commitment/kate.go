package commitment

import (
	"fmt"
	"math/big"

	"../../utils/conv"
	"../../utils/ecparam"
	"../../utils/polyring"
	. "github.com/Nik-U/pbc"
	. "github.com/ncw/gmp"
)

type DLPolyCommit struct {
	pairing *Pairing
	pk      []*Power
	degree  int
	p       *Int
}

// Generate New G1
func (c *DLPolyCommit) NewG1() *Element {
	return c.pairing.NewG1()
}

//Generate New GT
func (c *DLPolyCommit) NewGT() *Element {
	return c.pairing.NewGT()
}

// polyEval sets res to polyring(x)
func (c *DLPolyCommit) polyEval(res *Int, poly polyring.Polynomial, x *Int) {

	poly.EvalMod(x, c.p, res)
}

// Let polyring(x)=c0 + c1*x + ... cn * x^n, polyEvalInExponent sets res to g^polyring(alpha)
func (c *DLPolyCommit) polyEvalInExponent(res *Element, poly polyring.Polynomial) {
	// res = 1
	res.Set1()
	tmp := c.pairing.NewG1()
	for i := 0; i <= poly.GetDegree(); i++ {
		// tmp = g^{a^i} ^ ci
		ci, err := poly.GetCoefficient(i)
		if err != nil {
			panic("can't get coeff i")
		}

		c.pk[i].PowBig(tmp, conv.GmpInt2BigInt(&ci))
		res.Mul(res, tmp)
	}
}

// print the public keys
func (c *DLPolyCommit) printPublicKey() {
	for i := 0; i <= c.degree; i++ {
		fmt.Printf("g^(SK^%d): %s\n", i, c.pk[i].Source().String())
	}
}

var Curve = ecparam.PBC256

// SetupFix initializes a fixed pairing
func (c *DLPolyCommit) SetupFix(degree int) {
	c.degree = degree

	// setup the pairing
	c.pairing = Curve.Pairing
	c.p = Curve.Ngmp

	// trusted setup
	c.pk = make([]*Power, degree+1)

	// a generator g
	g := Curve.G

	// secret key
	sk := new(big.Int)
	sk.SetString("2", 10)

	tmp := new(big.Int)
	for i := 0; i <= degree; i++ {
		// tmp = sk ^ i
		bigP := big.NewInt(0)
		bigP.SetString(c.p.String(), 10)
		tmp.Exp(sk, big.NewInt(int64(i)), bigP)
		// pk[i] = g ^ tmp
		// Search pk and modify them all
		inter := c.pairing.NewG1()
		c.pk[i] = inter.PowBig(g, tmp).PreparePower()
	}
}

// Commit sets res to g^polyring(alpha)
func (c *DLPolyCommit) Commit(res *Element, poly polyring.Polynomial) {
	c.polyEvalInExponent(res, poly)
}

// Open is not used
func (c *DLPolyCommit) Open() {
	panic("unimplemented")
}

// VerifyPoly checks C == g ^ polyring(alpha)
func (c *DLPolyCommit) VerifyPoly(C *Element, poly polyring.Polynomial) bool {
	tmp := c.pairing.NewG1()
	c.polyEvalInExponent(tmp, poly)
	return tmp.Equals(C)
}

// CreateWitness sets res to g ^ phi(alpha) where phi(x) = (polyring(x)-polyring(x0)) / (x - x0)
func (c *DLPolyCommit) CreateWitness(res *Element, polynomial polyring.Polynomial, x0 *Int) {
	poly_t := polynomial.DeepCopy()

	// tmp = polynomial(x0)
	tmp := new(Int)
	c.polyEval(tmp, poly_t, x0)
	// fmt.Printf("CreateWitness\n%s\n%s\n", polynomial.String(), tmp.String())

	// poly_t = polynomial(x)-polynomial(x0)
	poly_t.GetPtrToConstant().Sub(poly_t.GetPtrToConstant(), tmp)

	// quot == poly_t / (x - x0)
	quot := polyring.NewEmpty()

	// denominator = x - x0
	denominator, err := polyring.New(1)
	if err != nil {
		panic("can't create polyring")
	}
	// FIXME: converting to int64 is dangerous
	denominator.SetCoefficient(1, 1)
	denominator.GetPtrToConstant().Neg(x0)

	quot.Div2(poly_t, denominator)
	// fmt.Printf("CreateWitness2\n%s\n", quot.String())

	c.polyEvalInExponent(res, quot)
}

// VerifyEval checks the correctness of w, returns true/false
func (c *DLPolyCommit) VerifyEval(C *Element, x *Int, polyX *Int, w *Element) bool {
	e1 := c.pairing.NewGT()
	e2 := c.pairing.NewGT()
	t1 := c.pairing.NewGT()
	t2 := c.pairing.NewG1()
	e1.Pair(C, c.pk[0].Source())
	exp := big.NewInt(0)
	exp.SetString(x.String(), 10)
	c.pk[0].PowBig(t2, exp)
	t2.Div(c.pk[1].Source(), t2)
	e2.Pair(w, t2)
	t1.Pair(c.pk[0].Source(), c.pk[0].Source())
	exp.SetString(polyX.String(), 10)
	t1.PowBig(t1, exp)
	e2.Mul(e2, t1)
	// fmt.Printf("e1\n%s\ne2\n%s\n", e1.String(), e2.String())
	return e1.Equals(e2)
}
