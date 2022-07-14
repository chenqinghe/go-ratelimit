package ratelimit

import (
	"container/ring"
)

type Ring struct {
	head, tail *ring.Ring
	full       bool
}

func (r *Ring) Forward(n int64) {
	for i := int64(0); i < n; i++ {
		if r.full {
			panic("head cursor moving too fast")
		}
		r.head = r.head.Move(1)
		if r.head == r.tail {
			r.full = true
		}
	}
}

func (r *Ring) Step(n int64) {
	for i := int64(0); i < n; i++ {
		if !r.full && r.tail == r.head {
			panic("tail cursor moving too fast")
		}
		r.tail = r.tail.Move(1)
		r.full = false
	}
}

func (r *Ring) TailElement() interface{} {
	if r.IsEmpty() {
		return nil
	}
	return r.tail.Value
}

func (r *Ring) HeadElement() interface{} {
	if r.IsEmpty() {
		return nil
	}
	return r.head.Prev().Value
}

func (r *Ring) EachElement(fn func(e interface{})) {
	if r.IsEmpty() {
		return
	}

	if r.full {
		fn(r.tail.Value)
		for p := r.tail.Next(); p != r.head; p = p.Next() {
			fn(p.Value)
		}
	} else {
		for p := r.tail; p != r.head; p = p.Next() {
			fn(p.Value)
		}
	}
}

func (r *Ring) IsEmpty() bool {
	return !r.full && r.head == r.tail
}

func NewRing(elements []interface{}) *Ring {
	if len(elements) == 0 {
		return nil
	}
	r := ring.New(len(elements))
	r.Value = elements[0]
	i := 1
	for p := r.Next(); p != r; p = p.Next() {
		p.Value = elements[i]
		i++
	}
	return &Ring{
		head: r,
		tail: r,
	}
}
