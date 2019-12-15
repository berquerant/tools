/*
Package functions provides utilities for functional programming
*/
package functions

import (
	"fmt"
	"reflect"
	"sort"
	"tools/pkg/conv/reflection"
	"tools/pkg/errors"
	"tools/pkg/functions/consume"
	"tools/pkg/functions/filter"
	"tools/pkg/functions/fold"
	"tools/pkg/functions/iterator"
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
		Map(mapper interface{}) Stream
		// Filter select elements
		//
		// predicate :: a -> bool
		Filter(predicate interface{}) Stream
		// Fold aggregate elements
		Fold(aggregator interface{}, options ...fold.FoldOption) Stream
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
		// Flat flatten stream, single level
		Flat() Stream
		// Lift lift up stream, single level, into []interface{}
		Lift() Stream
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
	errMsgCannotGetSlice       = "cannot get slice"
	errMsgCannotCompare        = "cannot compare"
	errMsgCannotConvert        = "cannot convert"
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

func (s *stream) Map(mapperFunc interface{}) Stream {
	f, err := mapper.NewMapper(mapperFunc)
	if err != nil {
		return NewNilStream(newStreamError(errors.Map, errMsgInvalidFunction, err))
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

func (s *stream) Filter(predicateFunc interface{}) Stream {
	f, err := filter.NewPredicate(predicateFunc)
	if err != nil {
		return NewNilStream(newStreamError(errors.Filter, errMsgInvalidFunction, err))
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

func (s *stream) Fold(aggregator interface{}, options ...fold.FoldOption) Stream {
	var err error
	f, err := fold.NewAggregator(aggregator)
	if err != nil {
		return NewNilStream(newStreamError(errors.Fold, errMsgInvalidFunction, err))
	}
	foldExecutor, err := fold.NewFoldExecutor(f, s, options...)
	if err != nil {
		return NewNilStream(newStreamError(errors.Fold, errMsgCannotCreateExecutor, err))
	}

	ret, err := foldExecutor.Fold()
	if err != nil {
		return NewNilStream(newStreamError(errors.Fold, errMsgCannotExecute, err))
	}
	return NewStream(iterator.MustNewFromInterfaces(ret))
}

func (s *stream) Consume(consumer interface{}) error {
	f, err := consume.NewConsumer(consumer)
	if err != nil {
		return newStreamError(errors.Consume, errMsgInvalidFunction, err)
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
	f, err := sorter.NewSorter(less)
	if err != nil {
		return NewNilStream(newStreamError(errors.Sort, errMsgInvalidFunction, err))
	}
	slice, err := iterator.ToSlice(s)
	if err != nil {
		return NewNilStream(newStreamError(errors.Sort, errMsgCannotGetSlice, err))
	}
	if len(slice) == 0 {
		return NewNilStream(nil)
	}
	var sortError error
	sort.SliceStable(slice, func(i, j int) bool {
		ret, err := f.Apply(slice[i], slice[j])
		if err != nil && sortError == nil {
			sortError = err
		}
		return ret
	})
	if sortError != nil {
		return NewNilStream(newStreamError(errors.Sort, errMsgCannotCompare, sortError))
	}
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

func getCommonType(v []interface{}) reflect.Type {
	defaultType := reflect.TypeOf([]interface{}{})
	if len(v) == 0 {
		return defaultType
	}
	t := reflect.TypeOf(v[0])
	for _, x := range v {
		if reflect.TypeOf(x).String() != t.String() {
			return defaultType
		}
	}
	return t
}

func (s *stream) Lift() Stream {
	var err error
	slice, err := iterator.ToSlice(s)
	if err != nil {
		return NewNilStream(newStreamError(errors.Lift, errMsgCannotGetSlice, err))
	}
	if len(slice) == 0 {
		return NewNilStream(nil)
	}

	t := getCommonType(slice)
	newSlice, err := reflection.Convert(slice, reflect.SliceOf(t))
	if err != nil {
		return NewNilStream(newStreamError(errors.Lift, errMsgCannotConvert, err))
	}

	var isYielded bool
	return NewStream(iterator.MustNew(iterator.Func(func() (interface{}, error) {
		if isYielded {
			return nil, iterator.EOI
		}
		isYielded = true
		return newSlice.Interface(), nil
	})))
}
