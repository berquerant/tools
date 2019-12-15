package consume

import (
	"fmt"
	"reflect"
	"tools/pkg/conv/reflection"
	"tools/pkg/errors"
)

var (
	InvalidConsumer = errors.NewError().SetCode(errors.Validate).SetError(fmt.Errorf("invalid consumer"))
)

type (
	// Consumer :: a
	Consumer interface {
		Apply(v interface{}) error
	}

	consumer struct {
		f interface{}
		v reflect.Value
		t reflect.Type
	}
)

func IsConsumer(f interface{}) bool {
	t := reflect.TypeOf(f)
	return t.Kind() == reflect.Func &&
		t.NumIn() == 1 && t.NumOut() == 0
}

func NewConsumer(f interface{}) (Consumer, errors.Error) {
	if !IsConsumer(f) {
		return nil, InvalidConsumer
	}
	return &consumer{
		f: f,
		v: reflect.ValueOf(f),
		t: reflect.TypeOf(f),
	}, nil
}

func (s *consumer) Apply(v interface{}) error {
	av, err := reflection.ConvertShallow(v, s.t.In(0))
	if err != nil {
		return errors.NewError().SetCode(errors.Conversion).SetError(fmt.Errorf("invalid argument for consumer: %v", err))
	}
	s.v.Call([]reflect.Value{av})
	return nil
}
