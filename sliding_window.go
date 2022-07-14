package ratelimit

import (
	"sync"
	"time"
)

type RingBuffer interface {
	Forward(n int64)
	Step(n int64)
	IsEmpty() bool
	TailElement() interface{}
	HeadElement() interface{}
	EachElement(fn func(e interface{}))
}

type Bucket struct {
	timestamp time.Time
	count     int64
}

func (b *Bucket) reset() {
	b.timestamp = time.Time{}
	b.count = 0
}

type SlidingWindow struct {
	mu *sync.Mutex
	r  RingBuffer

	size      time.Duration // 滑动窗口的总时间宽度
	precision time.Duration // 滑动窗口的精度，即每个bucket的时间宽度
}

func NewSlidingWindow(size, precision time.Duration) *SlidingWindow {
	count := int(size/precision) + 1
	elements := make([]interface{}, count)
	for i := 0; i < count; i++ {
		elements[i] = &Bucket{}
	}
	ring := NewRing(elements)

	return &SlidingWindow{
		mu:        &sync.Mutex{},
		r:         ring,
		size:      size,
		precision: precision,
	}
}

func (w *SlidingWindow) timeInBucket(moment time.Time) *Bucket {
	w.mu.Lock()
	defer w.mu.Unlock()

	// delete expired buckets
	for !w.r.IsEmpty() {
		bucket := w.r.TailElement().(*Bucket)
		if !bucket.timestamp.IsZero() && moment.Sub(bucket.timestamp) < w.size {
			break
		}
		bucket.reset()
		w.r.Step(1)
	}

	if w.r.IsEmpty() {
		w.r.Forward(1)
		b := w.r.HeadElement().(*Bucket)
		b.timestamp = moment
		return b
	}

	// determine bucket index
	num := moment.Sub(w.r.HeadElement().(*Bucket).timestamp) / w.precision // 差了多少个bucket
	w.r.Forward(int64(num))

	bucket := w.r.HeadElement().(*Bucket)
	bucket.timestamp = moment

	return bucket
}

func (w *SlidingWindow) CurrentBucket() *Bucket {
	return w.timeInBucket(time.Now())
}

func (w *SlidingWindow) TotalInDuration(d time.Duration) int64 {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.r.IsEmpty() {
		return 0
	}

	var total int64
	now := time.Now()
	w.r.EachElement(func(e interface{}) {
		b := e.(*Bucket)
		if now.Sub(b.timestamp) < d {
			total += b.count
		}
	})

	return total
}
