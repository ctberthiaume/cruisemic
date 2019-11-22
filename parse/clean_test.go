package parse

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhitelistNoop(t *testing.T) {
	assert := assert.New(t)
	b := []byte("hello world")
	n := len(b)
	bcopy := make([]byte, n)
	_ = copy(bcopy, b)
	nclean := Whitelist(b, n)
	assert.Equal(b[:nclean], bcopy)
	assert.Equal(n, nclean)
}

func TestWhitelistRemoveChars(t *testing.T) {
	assert := assert.New(t)
	b := []byte("hello world")
	b[0] = 3    // first
	b[5] = 0    // middle
	b[10] = 255 // last
	n := len(b)
	nclean := Whitelist(b, n)
	assert.Equal(n-3, nclean)
	assert.Equal([]byte{101, 108, 108, 111, 119, 111, 114, 108}, b[:nclean])
}

func TestWhitelistAllChars(t *testing.T) {
	assert := assert.New(t)
	n := 256
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = byte(i)
	}
	nclean := Whitelist(b, n)
	expected := make([]byte, ((126-32)+1)+3)
	expected[0] = byte(9)
	expected[1] = byte(10)
	expected[2] = byte(13)
	i := 3
	for char := 32; char <= 126; char++ {
		expected[i] = byte(char)
		i++
	}
	assert.Equal(expected, b[:nclean])
	assert.Equal(len(expected), nclean)
}

var nclean int
var clean []byte

func TestWhitelistEmpty(t *testing.T) {
	assert := assert.New(t)
	b := []byte("")
	n := len(b)
	nclean := Whitelist(b, n)
	assert.Equal(0, nclean)
	assert.Equal([]byte{}, b[:nclean])
}
