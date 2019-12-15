package sorter

import (
	"fmt"
	"reflect"
	"tools/pkg/conv/reflection"
	"tools/pkg/errors"
)

var (
	InvalidSorter = errors.NewError().SetCode(errors.Validate).SetError(fmt.Errorf("invalid sorter"))
)

type (
	// Sorter :: a -> a -> bool
	Sorter interface {
		Apply(x, y interface{}) (bool, error)
	}

	sorter struct {
		f interface{}
		v reflect.Value
		t reflect.Type
	}
)

func IsSorter(f interface{}) bool {
	t := reflect.TypeOf(f)
	return t.Kind() == reflect.Func &&
		t.NumIn() == 2 && t.NumOut() == 1 &&
		t.In(0).String() == t.In(1).String() &&
		t.Out(0).Kind() == reflect.Bool
}

func NewSorter(f interface{}) (Sorter, errors.Error) {
	if !IsSorter(f) {
		return nil, InvalidSorter
	}
	return &sorter{
		f: f,
		v: reflect.ValueOf(f),
		t: reflect.TypeOf(f),
	}, nil
}

func (s *sorter) Apply(x, y interface{}) (bool, error) {
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
		return false, errors.NewError().SetCode(errors.Conversion).SetError(fmt.Errorf("invalid argument for sorter: %v", err))
	}
	return s.v.Call([]reflect.Value{vx, vy})[0].Bool(), nil
}
