package functions

import (
	"fmt"
	"reflect"
	"tools/pkg/conv/reflection"
	"tools/pkg/errors"
)

var (
	InvalidAggregator = errors.NewError().SetCode(errors.Validate).SetError(fmt.Errorf("invalid aggregator"))
)

type (
	// Aggregator :: a -> b -> b
	Aggregator interface {
		Apply(x, acc interface{}) (interface{}, error)
	}

	aggregator struct {
		f interface{}
		v reflect.Value
		t reflect.Type
	}
)

func IsAggregator(f interface{}) bool {
	t := reflect.TypeOf(f)
	return t.Kind() == reflect.Func &&
		t.NumIn() == 2 && t.NumOut() == 1 &&
		t.In(1).String() == t.Out(0).String()
}

func NewAggregator(f interface{}) (Aggregator, errors.Error) {
	if !IsAggregator(f) {
		return nil, InvalidAggregator
	}
	return &aggregator{
		f: f,
		v: reflect.ValueOf(f),
		t: reflect.TypeOf(f),
	}, nil
}

func (s *aggregator) Apply(x, acc interface{}) (interface{}, error) {
	var (
		vx, vacc reflect.Value
	)
	if err := func() error {
		var err error
		if vx, err = reflection.Convert(x, reflect.Zero(s.t.In(0))); err != nil {
			return err
		}
		vacc, err = reflection.Convert(acc, reflect.Zero(s.t.In(1)))
		return err
	}(); err != nil {
		return nil, errors.NewError().SetCode(errors.Conversion).SetError(fmt.Errorf("invalid argument for aggregate: %v", err))
	}
	r := s.v.Call([]reflect.Value{vx, vacc})
	return r[0].Interface(), nil
}
