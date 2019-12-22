package iterator

// Join merges 2 iterators
func Join(x, y Iterator) Iterator {
	var useSecond bool
	r, _ := newIteratorFromFunc(func() (interface{}, error) {
		if useSecond {
			return y.Next()
		}
		elem, err := x.Next()
		if err == EOI {
			useSecond = true
			return y.Next()
		}
		return elem, err
	})
	return r
}

// ToReversed converts an iterator into an iterator that yields elements arranged in the reverse order.
// Original iterator must be finite.
func ToReversed(iter Iterator) (Iterator, error) {
	slice, err := ToSlice(iter)
	if err != nil {
		return nil, err
	}
	r := make([]interface{}, len(slice))
	for i, x := range slice {
		r[len(slice)-1-i] = x
	}
	return New(r)
}

type (
	cyclicIterator struct {
		idx   int
		slice []interface{}
	}
)

// ToCyclic converts an iterator into an infinite iterator that yields elements cyclicly.
// Original iterator must be finite and have an element at least.
func ToCyclic(iter Iterator) (Iterator, error) {
	slice, err := ToSlice(iter)
	if err != nil {
		return nil, err
	}
	if len(slice) == 0 {
		return nil, badIterator
	}
	return &cyclicIterator{
		slice: slice,
	}, nil
}

func (s *cyclicIterator) Next() (interface{}, error) {
	if s.idx >= len(s.slice) {
		s.idx = 0
	}
	v := s.slice[s.idx]
	s.idx++
	return v, nil
}

type (
	rangeIterator struct {
		start      int
		stop       int
		step       int
		isInfinite bool
		value      int
	}
	RangeIteratorBuider struct {
		start      int
		stop       int
		step       int
		isInfinite bool
	}
)

// NewRangeIteratorBuilder returns a builder of range iterator
func NewRangeIteratorBuilder() *RangeIteratorBuider {
	return &RangeIteratorBuider{
		start:      0,
		step:       1,
		stop:       0,
		isInfinite: false,
	}
}

// Start sets start value.
// default: 0
func (s *RangeIteratorBuider) Start(v int) *RangeIteratorBuider {
	s.start = v
	return s
}

// Stop sets stop value.
// it must be set when not infinite.
// range iterator will stop iteration when its value is over stop value.
func (s *RangeIteratorBuider) Stop(v int) *RangeIteratorBuider {
	s.stop = v
	return s
}

// Step sets step value.
// default: 1.
// range iterator yields int value increasing value by step
func (s *RangeIteratorBuider) Step(v int) *RangeIteratorBuider {
	s.step = v
	return s
}

// Infinite determines whether infinite iterator or not.
// Stop will be ignored when infinite iterator.
func (s *RangeIteratorBuider) Infinite(v bool) *RangeIteratorBuider {
	s.isInfinite = v
	return s
}

func (s *RangeIteratorBuider) Build() Iterator {
	return &rangeIterator{
		start:      s.start,
		stop:       s.stop,
		step:       s.step,
		isInfinite: s.isInfinite,
		value:      s.start,
	}
}

func (s *rangeIterator) Next() (interface{}, error) {
	v := s.value
	if s.isInfinite {
		s.value += s.step
		return v, nil
	}
	if v < s.stop {
		s.value += s.step
		return v, nil
	}
	return nil, EOI
}
