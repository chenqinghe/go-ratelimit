package ratelimit

import (
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type logger interface {
	Printf(format string, args ...interface{}) (n int, err error)
}

var Logger logger

type RateLimiter struct {
	mu *sync.Mutex

	metrics map[string]*SlidingWindow

	rules     []Rule
	precision time.Duration
}

type Rule struct {
	Duration time.Duration
	Limit    int64
	Level    Level
}

type Level int

func (rl *RateLimiter) Trigger(name string) Level {
	rl.mu.Lock()

	window, ok := rl.metrics[name]
	if !ok {
		rule := rl.rules[len(rl.rules)-1]
		window = NewSlidingWindow(rule.Duration, rl.precision)
		rl.metrics[name] = window
	}
	rl.mu.Unlock()

	bucket := window.CurrentBucket()
	atomic.AddInt64(&bucket.count, 1)

	var level Level
	for _, v := range rl.rules {
		total := window.TotalInDuration(v.Duration)
		if Logger != nil {
			Logger.Printf("level: %d, limit: %d, current: %d\n", v.Level, v.Limit, total)
		}
		if total >= v.Limit {
			level = v.Level
		}
	}

	return level
}

type NewRateLimiterOption struct {
	IgnoreRuleError bool
}

var defaultOption = &NewRateLimiterOption{
	IgnoreRuleError: false,
}

func NewRateLimiter(rules []Rule, precision time.Duration, opt ...*NewRateLimiterOption) (*RateLimiter, error) {
	sort.SliceStable(rules, func(i, j int) bool {
		return rules[i].Level < rules[j].Level
	})

	var option = defaultOption
	if len(opt) > 0 {
		option = opt[0]
	}
	if !option.IgnoreRuleError {
		if err := checkRule(rules); err != nil {
			return nil, err
		}
	}

	return &RateLimiter{
		mu:        &sync.Mutex{},
		metrics:   make(map[string]*SlidingWindow),
		rules:     rules,
		precision: precision,
	}, nil
}

var (
	ErrRuleLevelDuplicate   = errors.New("rule level can not be same")
	ErrRuleDurationImproper = errors.New("the higher the rule level, should have longer time duration")
	ErrRuleLimitImproper    = errors.New("the higher the rule level, should have greater limit threshold")
	ErrRuleQPSImproper      = errors.New("the higher the rule level, should have greater average qps")
)

// 限流规则约定：
//    1. 规则等级不能相同
//    2. 等级越高，时间跨度应该越大
//    3. 等级越高，时间跨度越大，则Limit也应该越大
//    3. 等级越高，平均qps应该越低，否则高等级规则永远不会触发
func checkRule(rules []Rule) error {
	if len(rules) == 1 {
		return nil
	}

	ruleMap := make(map[Level]struct{})
	for _, v := range rules {
		if _, ok := ruleMap[v.Level]; ok {
			return ErrRuleLevelDuplicate
		}
		ruleMap[v.Level] = struct{}{}
	}

	for i := 0; i < len(rules)-1; i++ {
		if rules[i].Duration >= rules[i+1].Duration {
			return ErrRuleDurationImproper
		}
	}

	for i := 0; i < len(rules)-1; i++ {
		if rules[i].Limit >= rules[i+1].Limit {
			return ErrRuleLimitImproper
		}
	}

	for i := 0; i < len(rules)-1; i++ {
		if float64(rules[i].Limit)/float64(rules[i].Duration) <= float64(rules[i+1].Limit)/float64(rules[i+1].Duration) {
			return ErrRuleQPSImproper
		}
	}

	return nil
}
