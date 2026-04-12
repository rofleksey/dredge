package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToPointer(t *testing.T) {
	s := "x"
	assert.Equal(t, &s, ToPointer("x"))

	b := true
	assert.Equal(t, &b, ToPointer(true))

	n := 42
	assert.Equal(t, &n, ToPointer(42))
}

func TestFromPointer(t *testing.T) {
	s := "hello"
	assert.Equal(t, "hello", FromPointer(&s))
	assert.Equal(t, "", FromPointer((*string)(nil)))

	b := true
	assert.Equal(t, true, FromPointer(&b))
	assert.Equal(t, false, FromPointer((*bool)(nil)))

	assert.Equal(t, 0, FromPointer((*int)(nil)))
}
