package encryption

import (
	"testing"

	"github.com/ncw/gmp"
	"github.com/stretchr/testify/assert"
	"math/rand"
)

func TestEncryption(t *testing.T) {

	pk := new(PublicKey)
	sk := new(PrivateKey)
	rnd := rand.New(rand.NewSource(99))
	q := gmp.NewInt(7)
	g := gmp.NewInt(2)

	KeyGen(pk, sk, q, g, rnd)
	m := gmp.NewInt(0)
	m.Rand(rnd, pk.q)

	ei := new(EncryptedInt)
	proof := new(Proof)
	Encrypt(ei, proof, pk, m, rnd)

	di := gmp.NewInt(0)
	Decrypt(di, pk, sk, ei)

	assert.Equal(t, m, di, "The decrypted message should be the same as the original message.")

	assert.True(t, Verify(proof, m, ei, pk), "Verify correct proof should give out true")
}
