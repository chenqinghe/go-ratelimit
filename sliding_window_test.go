package ratelimit

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewSlidingWindow(t *testing.T) {

	window := NewSlidingWindow(time.Second*10, time.Second)

	now := time.Now()

	for i := 0; i < 13; i++ {
		bucket := window.timeInBucket(now.Add(time.Second * time.Duration(i)))
		assert.Equal(t, time.Second*time.Duration(i), bucket.timestamp.Sub(now))
	}

}
