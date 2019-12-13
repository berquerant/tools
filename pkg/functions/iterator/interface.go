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

	errorIterator struct {
		e error
	}
	funcIterator struct {
		f Func
	}

	// KV is a cell to contain element from iterator made from map
	KV struct {
		K, V interface{}
	}
	ItemAndError struct {
		Item interface{}
		Err  error
	}
)

// ToChan converts iterator into channel.
// The channel is closed when iterator reached the end or some error
func ToChan(iter Iterator) (<-chan *ItemAndError, errors.Error) {
	ch := make(chan *ItemAndError)
	go func() {
		for {
			x, err := iter.Next()
			if err == EOI {
				close(ch)
				return
			}
			if err != nil {
				ch <- &ItemAndError{Err: err}
				close(ch)
				return
			}
			ch <- &ItemAndError{Item: x}
		}
	}()
	return ch, nil
}

// ToSlice convertes iterator into slice
func ToSlice(iter Iterator) ([]interface{}, errors.Error) {
	ret := []interface{}{}
	for {
		x, err := iter.Next()
		if err == EOI {
			return ret, nil
		}
		if err != nil {
			return nil, errors.NewError().SetCode(errors.Iterator).SetError(err)
		}
		ret = append(ret, x)
	}
}

// ToFunc converts iterator into function
func ToFunc(iter Iterator) (Func, errors.Error) {
	return Func(iter.Next), nil
}

func MustNew(v interface{}) Iterator {
	iter, err := New(v)
	if err != nil {
		panic(err)
	}
	return iter
}

func MustNewFromInterfaces(vs ...interface{}) Iterator {
	iter, err := NewFromInterfaces(vs...)
	if err != nil {
		panic(err)
	}
	return iter
}

func New(v interface{}) (Iterator, errors.Error) {
	if v == nil {
		return newNilIterator()
	}
	switch v := v.(type) {
	case Iterator:
		return v, nil
	case error:
		return newErrorIterator(v)
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
	}
	return nil, invalidArgument
}

func NewFromInterfaces(vs ...interface{}) (Iterator, errors.Error) {
	return newIteratorFromSlice(vs)
}

func newErrorIterator(err error) (Iterator, errors.Error) {
	if err == nil {
		return nil, invalidArgument
	}
	return &errorIterator{e: err}, nil
}

func (s *errorIterator) Next() (interface{}, error) {
	return nil, s.e
}

func newNilIterator() (Iterator, errors.Error) {
	return &errorIterator{e: EOI}, nil
}

func (s *funcIterator) Next() (interface{}, error) {
	return s.f()
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
			return &KV{
				K: iter.Key(),
				V: iter.Value(),
			}, nil
		}
		return nil, EOI
	}))
}
