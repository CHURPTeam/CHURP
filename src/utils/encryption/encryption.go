package encryption

import (
	"github.com/ncw/gmp"
	"crypto/sha256"
	"math/rand"
)

// We need a hash function

type PublicKey struct {
	q *gmp.Int
	g *gmp.Int
        h *gmp.Int
}

type PrivateKey struct {
	x *gmp.Int
}

type EncryptedInt struct {
        fst *gmp.Int
        snd *gmp.Int
}

type Proof struct {
	a *gmp.Int
	b *gmp.Int
	z *gmp.Int
}

func KeyGen(pk *PublicKey, sk *PrivateKey, q *gmp.Int, g *gmp.Int, rnd *rand.Rand) {

        pk.q = q
        pk.g = g
        sk.x = gmp.NewInt(0)
        sk.x.Rand(rnd, q)
        pk.h = gmp.NewInt(0)
        pk.h.Exp(g, sk.x, q)

}

func Encrypt(ei *EncryptedInt, proof *Proof, pk *PublicKey, m *gmp.Int, rnd *rand.Rand) {

        y := gmp.NewInt(0)
	y.Rand(rnd, pk.q)

	ei.fst = gmp.NewInt(0)
	ei.fst.Exp(pk.g, y, pk.q)
	ei.snd = gmp.NewInt(0)
	ei.snd.Exp(pk.h, y, pk.q)
	ei.snd.Mul(m, ei.snd)

        // pk.g, pk.h, ei.fst, ei.snd/m is a DH tuple
        r := gmp.NewInt(0)
	r.Rand(rnd, pk.q)
	proof.a = gmp.NewInt(0)
        proof.a.Exp(pk.g, r, pk.q)
	proof.b = gmp.NewInt(0)
	proof.b.Exp(pk.h, r, pk.q)
	bytes := append(pk.g.Bytes(), pk.h.Bytes()...)
	bytes = append(bytes, ei.fst.Bytes()...)
	tmp := gmp.NewInt(0)
	tmp.Div(ei.snd, m)
	bytes = append(bytes, tmp.Bytes()...)
	e := gmp.NewInt(0)
	tmp2 := sha256.Sum256(bytes)
	e.SetBytes(tmp2[:])
	proof.z = gmp.NewInt(0)
        proof.z.Mul(e, y)
	proof.z.Add(r, proof.z)
}

func Decrypt(di *gmp.Int, pk *PublicKey, sk *PrivateKey, ei *EncryptedInt) {

	di.Exp(ei.fst, sk.x, pk.q)
	di.Div(ei.snd, di)
}

func Verify(proof *Proof, m *gmp.Int, ei *EncryptedInt, pk *PublicKey) bool {

        bytes := append(pk.g.Bytes(), pk.h.Bytes()...)
        bytes = append(bytes, ei.fst.Bytes()...)
        tmp := gmp.NewInt(0)
        tmp.Div(ei.snd, m)
        bytes = append(bytes, tmp.Bytes()...)
        e := gmp.NewInt(0)

	tmp1 := gmp.NewInt(0)
	tmp1.Exp(pk.g, proof.z, pk.q)

	tmp2 := gmp.NewInt(0)
	tmp2.Exp(ei.fst, e, pk.q)
	tmp2.Mul(proof.a, tmp2)

	judgement := (tmp1.Cmp(tmp2) == 0)

	tmp1.Exp(pk.h, proof.z, pk.q)

        tmp2.Exp(tmp, e, pk.q)
	tmp2.Mul(proof.b, tmp2)

	judgement = judgement && (tmp1.Cmp(tmp2) == 0)

	return judgement
}
