package functions

import (
	"fmt"
	"tools/pkg/errors"
	"tools/pkg/functions/iterator"
)

var (
	InvalidResult = errors.NewError().SetCode(errors.Validate).SetError(fmt.Errorf("invalid result"))
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
		// Consum consume stream
		//
		// consumer :: a
		Consume(consumer interface{}) error
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
		panic(err)
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
		panic(err)
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
	f, err := NewAggregator(aggregator)
	if err != nil {
		panic(err)
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

	iter := func() iterator.Iterator {
		ret, err := foldr(f, initialValue, s)
		if err != nil {
			return iterator.MustNew(err)
		}
		return iterator.MustNewFromInterfaces(ret)
	}()
	return NewStream(iter)
}

func (s *stream) Consume(consumer interface{}) error {
	f, err := NewConsumer(consumer)
	if err != nil {
		panic(err)
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
