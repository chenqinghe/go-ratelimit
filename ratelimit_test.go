package ratelimit

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type stdlog struct{}

func (l stdlog) Printf(format string, args ...interface{}) (int, error) {
	return fmt.Printf(format, args...)
}

func TestCheckRule(t *testing.T) {

	err1 := checkRule([]Rule{
		{
			Level:    1,
			Duration: time.Second,
		},
		{
			Level:    1,
			Duration: time.Second,
		},
	})
	assert.Equal(t, ErrRuleLevelDuplicate, err1)
	err2 := checkRule([]Rule{
		{
			Level:    1,
			Duration: time.Second,
			Limit:    10,
		},
		{
			Level:    2,
			Duration: time.Second,
			Limit:    10,
		},
		{
			Level:    3,
			Duration: time.Second * 3,
			Limit:    10,
		},
	})
	assert.Equal(t, ErrRuleDurationImproper, err2)

	err3 := checkRule([]Rule{
		{
			Level:    1,
			Duration: time.Second,
			Limit:    10,
		},
		{
			Level:    2,
			Duration: time.Second * 2,
			Limit:    20,
		},
		{
			Level:    3,
			Duration: time.Second * 3,
			Limit:    10,
		},
	})
	assert.Equal(t, ErrRuleLimitImproper, err3)

	err4 := checkRule([]Rule{
		{
			Level:    1,
			Duration: time.Second,
			Limit:    10,
		},
		{
			Level:    2,
			Duration: time.Second * 2,
			Limit:    30,
		},
		{
			Level:    3,
			Duration: time.Second * 3,
			Limit:    100,
		},
	})
	assert.Equal(t, ErrRuleQPSImproper, err4)

	err5 := checkRule([]Rule{
		{
			Level:    1,
			Duration: time.Second,
			Limit:    10,
		},
		{
			Level:    2,
			Duration: time.Second * 2,
			Limit:    15,
		},
		{
			Level:    3,
			Duration: time.Second * 3,
			Limit:    20,
		},
	})

	assert.Equal(t, nil, err5)
}

func TestRateLimiter_Trigger2(t *testing.T) {
	Logger = stdlog{}
	ratelimit, _ := NewRateLimiter([]Rule{
		{
			Duration: time.Second,
			Limit:    100,
			Level:    1,
		},
		{
			Duration: time.Second * 10,
			Limit:    500,
			Level:    2,
		},
	}, time.Millisecond*10)

	wg := &sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()

			now := time.Now()
			for {
				lv := ratelimit.Trigger("hello")
				if lv > 0 {
					fmt.Println("current level:", lv)
					return
				}
				time.Sleep(time.Millisecond * 100)
				if time.Now().Sub(now) > time.Second*12 {
					return
				}
			}

		}()

		time.Sleep(time.Second * 2)
	}

	wg.Wait()

}

func TestRateLimiter_Trigger(t *testing.T) {
	Logger = stdlog{}
	ratelimit, _ := NewRateLimiter([]Rule{
		{
			Duration: time.Second,
			Limit:    100,
			Level:    1,
		},
		{
			Duration: time.Second * 10,
			Limit:    500,
			Level:    2,
		},
	}, time.Millisecond*10)

	ratelimit.Trigger("test")
	ratelimit.Trigger("test")
	ratelimit.Trigger("test")
	ratelimit.Trigger("test")
}

func BenchmarkRateLimit_Trigger(b *testing.B) {
	ratelimit, _ := NewRateLimiter([]Rule{
		{
			Duration: time.Second,
			Limit:    100,
			Level:    1,
		},
		{
			Duration: time.Second * 10,
			Limit:    500,
			Level:    2,
		},
	}, time.Millisecond*10)

	for i := 0; i < b.N; i++ {
		ratelimit.Trigger("test")
	}
}
