package fold

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
	// Aggregator :: a -> b -> b or b -> a -> b
	Aggregator interface {
		Apply(x, acc interface{}) (interface{}, error)
		Type() AggregatorType
		// IV returns initial zero value
		IV() interface{}
	}

	aggregator struct {
		f  interface{}
		v  reflect.Value
		t  reflect.Type
		at AggregatorType
	}
)

//go:generate stringer -type=AggregatorType -output generated.aggregatortype_string.go
type AggregatorType int

const (
	UnknownAggregator AggregatorType = iota
	// RightAggregator :: a -> b -> b
	RightAggregator
	// LeftAggregator :: b -> a -> b
	LeftAggregator
	// PerfectAggregator :: a -> a -> a
	PerfectAggregator
)

// IsRightAggregator returns true if f :: a -> b -> b
func IsRightAggregator(f interface{}) bool {
	t := reflect.TypeOf(f)
	return t.Kind() == reflect.Func &&
		t.NumIn() == 2 && t.NumOut() == 1 &&
		t.In(1).String() == t.Out(0).String()
}

// IsLeftAggregator returns true if f :: b -> a -> b
func IsLeftAggregator(f interface{}) bool {
	t := reflect.TypeOf(f)
	return t.Kind() == reflect.Func &&
		t.NumIn() == 2 && t.NumOut() == 1 &&
		t.In(0).String() == t.Out(0).String()
}

// IsPerfectAggregator returns true if f :: a -> a -> a
func IsPerfectAggregator(f interface{}) bool {
	return IsRightAggregator(f) && IsLeftAggregator(f)
}

var (
	aggregatorTypeTuples = []struct {
		Tester func(interface{}) bool
		Type   AggregatorType
	}{
		{
			Tester: IsPerfectAggregator,
			Type:   PerfectAggregator,
		},
		{
			Tester: IsRightAggregator,
			Type:   RightAggregator,
		},
		{
			Tester: IsLeftAggregator,
			Type:   LeftAggregator,
		},
	}
)

func GetAggregatorType(f interface{}) AggregatorType {
	for _, t := range aggregatorTypeTuples {
		if t.Tester(f) {
			return t.Type
		}
	}
	return UnknownAggregator
}

func IsAggregator(f interface{}) bool {
	return IsRightAggregator(f) || IsLeftAggregator(f)
}

func NewAggregator(f interface{}) (Aggregator, errors.Error) {
	if !IsAggregator(f) {
		return nil, InvalidAggregator
	}
	return &aggregator{
		f:  f,
		v:  reflect.ValueOf(f),
		t:  reflect.TypeOf(f),
		at: GetAggregatorType(f),
	}, nil
}

func (s *aggregator) IV() interface{} {
	return reflect.Zero(s.t.Out(0)).Interface()
}

func (s *aggregator) Type() AggregatorType {
	return s.at
}

func (s *aggregator) Apply(x, y interface{}) (interface{}, error) {
	var (
		vx, vy reflect.Value
	)
	if err := func() error {
		var err error
		if vx, err = reflection.ConvertShallow(x, s.t.In(0)); err != nil {
			return err
		}
		vy, err = reflection.ConvertShallow(y, s.t.In(1))
		return err
	}(); err != nil {
		return nil, errors.NewError().SetCode(errors.Conversion).SetError(fmt.Errorf("invalid argument for aggregate: %v", err))
	}
	r := s.v.Call([]reflect.Value{vx, vy})
	return r[0].Interface(), nil
}
