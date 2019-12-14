/*
Package functions provides utilities for functional programming
*/
package functions

import (
	"fmt"
	"reflect"
	"sort"
	"tools/pkg/conv/reflection"
	"tools/pkg/functions/iterator"
)

type (
	// Stream provides higher order functions
	Stream interface {
		iterator.Iterator
		// Map convert each elements
		//
		// mapper :: a -> b
		Map(mapper interface{}) Stream
		// Filter select elements
		//
		// predicate :: a -> bool
		Filter(predicate interface{}) Stream
		// Fold aggregate elements
		//
		// aggregator :: a -> b -> b
		//
		// initialValue :: b
		Fold(aggregator, initialValue interface{}) Stream
		// Consume consume stream
		//
		// consumer :: a
		Consume(consumer interface{}) error
		// As assign stream into reference v
		As(v interface{}) error
		// Sort sort stream
		//
		// less :: a -> a -> bool
		Sort(less interface{}) Stream
		// Flat flatten stream
		Flat() Stream
	}

	stream struct {
		iter iterator.Iterator
	}
)

func NewStream(iter iterator.Iterator) Stream {
	return &stream{iter: iter}
}

func (s *stream) Next() (interface{}, error) {
	return s.iter.Next()
}

func (s *stream) Map(mapper interface{}) Stream {
	f, err := NewMapper(mapper)
	if err != nil {
		panic(fmt.Sprintf("invalid function for Map: %v", err))
	}

	return NewStream(iterator.MustNew(iterator.Func(func() (interface{}, error) {
		x, err := s.Next()
		if err != nil {
			return nil, err
		}
		ret, err := f.Apply(x)
		if err != nil {
			return nil, err
		}
		return ret, nil
	})))
}

func (s *stream) Filter(predicate interface{}) Stream {
	f, err := NewPredicate(predicate)
	if err != nil {
		panic(fmt.Sprintf("invalid function for Filter: %v", err))
	}

	var iFunc func() (interface{}, error)
	iFunc = func() (interface{}, error) {
		x, err := s.Next()
		if err != nil {
			return nil, err
		}
		ret, err := f.Apply(x)
		if err != nil {
			return nil, err
		}
		if !ret {
			return iFunc()
		}
		return x, nil
	}
	return NewStream(iterator.MustNew(iterator.Func(iFunc)))
}

func (s *stream) Fold(aggregator, initialValue interface{}) Stream {
	var err error
	f, err := NewAggregator(aggregator)
	if err != nil {
		panic(fmt.Sprintf("invalid function for Fold: %v", err))
	}

	var foldr func(f Aggregator, acc interface{}, iter iterator.Iterator) (interface{}, error)
	foldr = func(f Aggregator, acc interface{}, iter iterator.Iterator) (interface{}, error) {
		x, err := iter.Next()
		if err == iterator.EOI {
			return acc, nil
		}
		if err != nil {
			return nil, err
		}
		ret, err := foldr(f, acc, iter)
		if err != nil {
			return nil, err
		}
		return f.Apply(x, ret)
	}

	ret, err := foldr(f, initialValue, s)
	if err != nil {
		panic(fmt.Sprintf("cannot fold for Fold: %v", err))
	}
	return NewStream(iterator.MustNewFromInterfaces(ret))
}

func (s *stream) Consume(consumer interface{}) error {
	f, err := NewConsumer(consumer)
	if err != nil {
		panic(fmt.Sprintf("invalid function for Consume: %v", err))
	}

	for {
		x, err := s.Next()
		if err == iterator.EOI {
			return nil
		}
		if err != nil {
			return err
		}
		if err := f.Apply(x); err != nil {
			return err
		}
	}
}

func (s *stream) As(v interface{}) error {
	slice, err := iterator.ToSlice(s)
	if err != nil {
		return err
	}
	sv, err := reflection.Convert(slice, reflect.TypeOf(v).Elem())
	if err != nil {
		return err
	}
	reflect.ValueOf(v).Elem().Set(sv)
	return nil
}

func (s *stream) Sort(less interface{}) Stream {
	var err error
	f, err := NewSorter(less)
	if err != nil {
		panic(fmt.Sprintf("invalid function for Sort: %v", err))
	}
	slice, err := iterator.ToSlice(s)
	if err != nil {
		panic(fmt.Sprintf("cannot get slice for Sort: %v", err))
	}
	if len(slice) == 0 {
		return NewStream(iterator.MustNew(nil))
	}
	sort.SliceStable(slice, func(i, j int) bool {
		ret, err := f.Apply(slice[i], slice[j])
		if err != nil {
			panic(fmt.Sprintf("cannot compare for Sort: %v", err))
		}
		return ret
	})
	return NewStream(iterator.MustNew(slice))
}

func (s *stream) Flat() Stream {
	var (
		top   iterator.Iterator
		iFunc func() (interface{}, error)
	)
	iFunc = func() (interface{}, error) {
		if top == nil {
			elem, err := s.Next()
			if err != nil {
				return nil, err
			}
			top = iterator.MustNew(elem)
		}

		x, err := top.Next()
		if err == iterator.EOI {
			top = nil
			return iFunc()
		}
		if err != nil {
			return nil, err
		}
		return x, nil
	}

	return NewStream(iterator.MustNew(iterator.Func(iFunc)))
}
