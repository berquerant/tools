package mapper

import (
	"fmt"
	"reflect"
	"tools/pkg/conv/reflection"
	"tools/pkg/errors"
)

var (
	InvalidMapper = errors.NewError().SetCode(errors.Validate).SetError(fmt.Errorf("invalid mapper"))
)

type (
	// Mapper :: a -> b
	Mapper interface {
		Apply(v interface{}) (interface{}, error)
	}

	mapper struct {
		f interface{}
		v reflect.Value
		t reflect.Type
	}
)

func IsMapper(f interface{}) bool {
	t := reflect.TypeOf(f)
	return t.Kind() == reflect.Func &&
		t.NumIn() == 1 && t.NumOut() == 1
}

func NewMapper(f interface{}) (Mapper, errors.Error) {
	if !IsMapper(f) {
		return nil, InvalidMapper
	}
	return &mapper{
		f: f,
		v: reflect.ValueOf(f),
		t: reflect.TypeOf(f),
	}, nil
}

func (s *mapper) Apply(v interface{}) (interface{}, error) {
	av, err := reflection.ConvertShallow(v, s.t.In(0))
	if err != nil {
		return nil, errors.NewError().SetCode(errors.Conversion).SetError(fmt.Errorf("invalid argument for mapper: %v", err))
	}
	r := s.v.Call([]reflect.Value{av})
	return r[0].Interface(), nil
}
