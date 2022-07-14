package ratelimit

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRing_Forward(t *testing.T) {
	elements := make([]interface{}, 10)
	for k := range elements {
		elements[k] = k + 1
	}
	ring := NewRing(elements)

	for i := 0; i < 10; i++ {
		ring.Forward(1)
		assert.Equal(t, elements[i], ring.HeadElement())
	}

	assert.Equal(t, true, ring.full)

	ring.EachElement(func(e interface{}) {
		fmt.Println(e)
	})

}
