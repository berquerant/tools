/*
Package iterator provides iterator
*/
package iterator

import (
	"fmt"
	"reflect"
	"tools/pkg/errors"
)

var (
	// EOI indicates end of interator
	EOI             = errors.NewError().SetCode(errors.Normal).SetError(fmt.Errorf("EOI"))
	invalidArgument = errors.NewError().SetCode(errors.Validate).SetError(fmt.Errorf("invalid argument"))
	badIterator     = errors.NewError().SetCode(errors.Iterator).SetError(fmt.Errorf("bad iterator"))
)

type (
	// Iterator is like a sequence
	Iterator interface {
		// Next returns next element
		Next() (interface{}, error)
	}
	// Func is an iterator as a function
	Func func() (interface{}, error)
)

func MustNew(v interface{}) Iterator {
	iter, err := New(v)
	if err != nil {
		panic(fmt.Sprintf("cannot create iterator: %v", err))
	}
	return iter
}

func MustNewFromInterfaces(vs ...interface{}) Iterator {
	iter, err := NewFromInterfaces(vs...)
	if err != nil {
		panic(fmt.Sprintf("cannot create iterator: %v", err))
	}
	return iter
}

// CanBeSingleValueIterator returns true if New(v) yields just an item
func CanBeSingleValueIterator(v interface{}) bool {
	if v == nil {
		return false
	}
	switch v.(type) {
	case Iterator, Func:
		return false
	}
	switch reflect.TypeOf(v).Kind() {
	case reflect.Array, reflect.Slice, reflect.Chan, reflect.Map:
		return false
	default:
		return true
	}
}

func New(v interface{}) (Iterator, errors.Error) {
	if v == nil {
		return newNilIterator()
	}
	switch v := v.(type) {
	case Iterator:
		return v, nil
	case Func:
		return newIteratorFromFunc(v)
	}
	switch reflect.TypeOf(v).Kind() {
	case reflect.Array:
		return newIteratorFromArray(v)
	case reflect.Slice:
		return newIteratorFromSlice(v)
	case reflect.Chan:
		return newIteratorFromChan(v)
	case reflect.Map:
		return newIteratorFromMap(v)
	default:
		return newSingleValueIterator(v)
	}
}

func NewFromInterfaces(vs ...interface{}) (Iterator, errors.Error) {
	return newIteratorFromSlice(vs)
}

func newSingleValueIterator(v interface{}) (Iterator, errors.Error) {
	if v == nil {
		return nil, invalidArgument
	}
	var isYielded bool
	return newIteratorFromFunc(func() (interface{}, error) {
		if isYielded {
			return nil, EOI
		}
		isYielded = true
		return v, nil
	})
}

func newNilIterator() (Iterator, errors.Error) {
	return newIteratorFromFunc(func() (interface{}, error) {
		return nil, EOI
	})
}

type (
	funcIterator struct {
		f     Func
		isEOI bool
	}
)

func (s *funcIterator) Next() (interface{}, error) {
	if s.isEOI {
		return nil, EOI
	}
	x, err := s.f()
	if err != nil {
		s.isEOI = true
		return nil, err
	}
	return x, nil
}

func newIteratorFromFunc(f Func) (Iterator, errors.Error) {
	if f == nil {
		return nil, invalidArgument
	}
	return &funcIterator{f: f}, nil
}

func newIteratorFromArray(v interface{}) (Iterator, errors.Error) {
	if v == nil {
		return nil, invalidArgument
	}
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Array {
		return nil, invalidArgument
	}
	array := reflect.ValueOf(v)
	var i int
	return newIteratorFromFunc(Func(func() (interface{}, error) {
		if i >= array.Len() {
			return nil, EOI
		}
		x := array.Index(i)
		i++
		return x.Interface(), nil
	}))
}

func newIteratorFromSlice(v interface{}) (Iterator, errors.Error) {
	if v == nil {
		return nil, invalidArgument
	}
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Slice {
		return nil, invalidArgument
	}
	slice := reflect.ValueOf(v)
	var i int
	return newIteratorFromFunc(Func(func() (interface{}, error) {
		if i >= slice.Len() {
			return nil, EOI
		}
		x := slice.Index(i)
		i++
		return x.Interface(), nil
	}))
}

func newIteratorFromChan(v interface{}) (Iterator, errors.Error) {
	if v == nil {
		return nil, invalidArgument
	}
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Chan || t.ChanDir() == reflect.SendDir {
		return nil, invalidArgument
	}
	c := reflect.ValueOf(v)
	return newIteratorFromFunc(Func(func() (interface{}, error) {
		x, ok := c.Recv()
		if ok {
			return x.Interface(), nil
		}
		return nil, EOI
	}))
}

type (
	// KV is a cell to contain element from iterator made from map
	KV interface {
		K() interface{}
		V() interface{}
	}
	kv struct {
		k, v interface{}
	}
)

func (s *kv) K() interface{} { return s.k }
func (s *kv) V() interface{} { return s.v }

func newIteratorFromMap(v interface{}) (Iterator, errors.Error) {
	if v == nil {
		return nil, invalidArgument
	}
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Map {
		return nil, invalidArgument
	}
	iter := reflect.ValueOf(v).MapRange()
	return newIteratorFromFunc(Func(func() (interface{}, error) {
		if iter.Next() {
			return &kv{
				k: iter.Key().Interface(),
				v: iter.Value().Interface(),
			}, nil
		}
		return nil, EOI
	}))
}
