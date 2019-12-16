/*
Package functions provides utilities for functional programming
*/
package functions

import (
	"fmt"
	"reflect"
	"tools/pkg/conv/reflection"
	"tools/pkg/errors"
	"tools/pkg/functions/consume"
	"tools/pkg/functions/filter"
	"tools/pkg/functions/flat"
	"tools/pkg/functions/fold"
	"tools/pkg/functions/iterator"
	"tools/pkg/functions/lift"
	"tools/pkg/functions/mapper"
	"tools/pkg/functions/sorter"
)

type (
	// Stream provides higher order functions
	Stream interface {
		iterator.Iterator
		// Map convert each elements
		//
		// mapper :: a -> b
		Map(mapper interface{}, options ...mapper.Option) Stream
		// Filter select elements
		//
		// predicate :: a -> bool
		Filter(predicate interface{}, options ...filter.Option) Stream
		// Fold aggregate elements
		Fold(aggregator interface{}, options ...fold.Option) Stream
		// Consume consume stream
		//
		// consumer :: a
		Consume(consumer interface{}, options ...consume.Option) error
		// As assign stream into reference v
		As(v interface{}) error
		// Sort sort stream
		//
		// less :: a -> a -> bool
		Sort(less interface{}, options ...sorter.Option) Stream
		// Flat flatten stream, single level
		Flat(options ...flat.Option) Stream
		// Lift lift up stream, single level, into []interface{}
		Lift(options ...lift.Option) Stream
		// Err get error during streaming.
		// should invoke before extracting result.
		// stream is nil stream when err is not nil
		Err() error
	}

	stream struct {
		iter iterator.Iterator
		err  error
	}
)

const (
	errMsgInvalidFunction      = "invalid function"
	errMsgCannotCreateExecutor = "cannot create executor"
	errMsgCannotExecute        = "cannot execute"
	errMsgCannotCompare        = "cannot compare"
)

func newStreamError(code errors.Code, msg string, err error) error {
	return errors.NewError().SetCode(code).SetError(fmt.Errorf("%s: %v", msg, err))
}

func NewStream(iter iterator.Iterator) Stream {
	return &stream{iter: iter}
}

// NewNilStream create stream that yield no items.
// Err() returns error if you set not nil error
func NewNilStream(err error) Stream {
	return &stream{
		iter: iterator.MustNew(nil),
		err:  err,
	}
}

func (s *stream) Next() (interface{}, error) {
	return s.iter.Next()
}

func (s *stream) Err() error {
	return s.err
}

func (s *stream) Map(mapperFunc interface{}, options ...mapper.Option) Stream {
	f, err := mapper.NewMapper(mapperFunc)
	if err != nil {
		return NewNilStream(newStreamError(errors.Map, errMsgInvalidFunction, err))
	}
	mapExecutor, err := mapper.NewExecutor(f, s, options...)
	if err != nil {
		return NewNilStream(newStreamError(errors.Map, errMsgCannotCreateExecutor, err))
	}
	return NewStream(mapExecutor.Execute())
}

func (s *stream) Filter(predicateFunc interface{}, options ...filter.Option) Stream {
	f, err := filter.NewPredicate(predicateFunc)
	if err != nil {
		return NewNilStream(newStreamError(errors.Filter, errMsgInvalidFunction, err))
	}
	filterExecutor, err := filter.NewExecutor(f, s, options...)
	if err != nil {
		return NewNilStream(newStreamError(errors.Filter, errMsgCannotCreateExecutor, err))
	}
	return NewStream(filterExecutor.Execute())
}

func (s *stream) Fold(aggregator interface{}, options ...fold.Option) Stream {
	var err error
	f, err := fold.NewAggregator(aggregator)
	if err != nil {
		return NewNilStream(newStreamError(errors.Fold, errMsgInvalidFunction, err))
	}
	foldExecutor, err := fold.NewExecutor(f, s, options...)
	if err != nil {
		return NewNilStream(newStreamError(errors.Fold, errMsgCannotCreateExecutor, err))
	}
	ret, err := foldExecutor.Execute()
	if err != nil {
		return NewNilStream(newStreamError(errors.Fold, errMsgCannotExecute, err))
	}
	return NewStream(iterator.MustNewFromInterfaces(ret))
}

func (s *stream) Consume(consumer interface{}, options ...consume.Option) error {
	f, err := consume.NewConsumer(consumer)
	if err != nil {
		return newStreamError(errors.Consume, errMsgInvalidFunction, err)
	}
	consumeExecutor, err := consume.NewExecutor(f, s, options...)
	if err != nil {
		return newStreamError(errors.Consume, errMsgCannotCreateExecutor, err)
	}
	return consumeExecutor.Execute()
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

func (s *stream) Sort(less interface{}, options ...sorter.Option) Stream {
	var err error
	f, err := sorter.NewSorter(less)
	if err != nil {
		return NewNilStream(newStreamError(errors.Sort, errMsgInvalidFunction, err))
	}
	sortExecutor, err := sorter.NewExecutor(f, s, options...)
	if err != nil {
		return NewNilStream(newStreamError(errors.Sort, errMsgCannotCreateExecutor, err))
	}
	iter, err := sortExecutor.Execute()
	if err != nil {
		return NewNilStream(newStreamError(errors.Sort, errMsgCannotCompare, err))
	}
	return NewStream(iter)
}

func (s *stream) Flat(options ...flat.Option) Stream {
	flatExecutor, err := flat.NewExecutor(s, options...)
	if err != nil {
		return NewNilStream(newStreamError(errors.Flat, errMsgCannotCreateExecutor, err))
	}
	return NewStream(flatExecutor.Execute())
}

func (s *stream) Lift(options ...lift.Option) Stream {
	var err error
	liftExecutor, err := lift.NewExecutor(s, options...)
	if err != nil {
		return NewNilStream(newStreamError(errors.Lift, errMsgCannotCreateExecutor, err))
	}
	iter, err := liftExecutor.Execute()
	if err != nil {
		return NewNilStream(newStreamError(errors.Lift, errMsgCannotExecute, err))
	}
	return NewStream(iter)
}
