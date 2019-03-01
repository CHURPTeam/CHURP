package commitment

import (
	"testing"

	"github.com/ncw/gmp"
	"github.com/stretchr/testify/assert"
)

func TestDLCommit(t *testing.T) {
	c := DLCommit{}
	c.SetupFix()

	// res = g^x
	res := c.pairing.NewG1()
	x := gmp.NewInt(100)
	c.Commit(res, x)

	assert.True(t, c.Verify(res, x), "dl_commit")
}
