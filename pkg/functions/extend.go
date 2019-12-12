package functions

import (
	"sort"
	"tools/pkg/functions/iterator"
)

type (
	ExtendedStream interface {
		iterator.Iterator
		Map(mapper interface{}) ExtendedStream
		Filter(predicate interface{}) ExtendedStream
		Fold(aggregator, initialValue interface{}) ExtendedStream
		Consume(consumer interface{}) error
		// Sort sorts stream
		//
		// less :: a -> a -> bool
		Sort(less interface{}) ExtendedStream
		Slice() ([]interface{}, error)
	}

	extendedStream struct {
		st Stream
	}
)

func NewExtendedStream(st Stream) ExtendedStream {
	return &extendedStream{st: st}
}

func (s *extendedStream) Map(mapper interface{}) ExtendedStream {
	return NewExtendedStream(s.st.Map(mapper))
}

func (s *extendedStream) Filter(predicate interface{}) ExtendedStream {
	return NewExtendedStream(s.st.Filter(predicate))
}

func (s *extendedStream) Fold(aggregator, initialValue interface{}) ExtendedStream {
	return NewExtendedStream(s.st.Fold(aggregator, initialValue))
}

func (s *extendedStream) Consume(consumer interface{}) error {
	return s.st.Consume(consumer)
}

func (s *extendedStream) Next() (interface{}, error) {
	return s.st.Next()
}

func (s *extendedStream) Slice() ([]interface{}, error) {
	return iterator.ToSlice(s)
}

func (s *extendedStream) Sort(less interface{}) ExtendedStream {
	var err error
	f, err := NewSorter(less)
	if err != nil {
		panic(err)
	}
	slice, err := s.Slice()
	if err != nil {
		panic(err)
	}
	sort.SliceStable(slice, func(i, j int) bool {
		ret, err := f.Apply(slice[i], slice[j])
		if err != nil {
			panic(err)
		}
		return ret
	})
	return NewExtendedStream(NewStream(iterator.MustNew(slice)))
}
