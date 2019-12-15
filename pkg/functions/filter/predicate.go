package filter

import (
	"fmt"
	"reflect"
	"tools/pkg/conv/reflection"
	"tools/pkg/errors"
)

var (
	InvalidPredicate = errors.NewError().SetCode(errors.Validate).SetError(fmt.Errorf("invalid predicate"))
)

type (
	// Predicate :: a -> bool
	Predicate interface {
		Apply(v interface{}) (bool, error)
	}

	predicate struct {
		f interface{}
		v reflect.Value
		t reflect.Type
	}
)

func IsPredicate(f interface{}) bool {
	t := reflect.TypeOf(f)
	return t.Kind() == reflect.Func &&
		t.NumIn() == 1 && t.NumOut() == 1 &&
		t.Out(0).Kind() == reflect.Bool
}

func NewPredicate(f interface{}) (Predicate, errors.Error) {
	if !IsPredicate(f) {
		return nil, InvalidPredicate
	}
	return &predicate{
		f: f,
		v: reflect.ValueOf(f),
		t: reflect.TypeOf(f),
	}, nil
}

func (s *predicate) Apply(v interface{}) (bool, error) {
	av, err := reflection.ConvertShallow(v, s.t.In(0))
	if err != nil {
		return false, errors.NewError().SetCode(errors.Conversion).SetError(fmt.Errorf("invalid argument for predicate: %v", err))
	}
	r := s.v.Call([]reflect.Value{av})
	return r[0].Bool(), nil
}
