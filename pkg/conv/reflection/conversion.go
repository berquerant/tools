/*
Package reflection provides reflection utilties
*/
package reflection

import (
	"fmt"
	"reflect"
	"tools/pkg/errors"
)

var (
	shallowDefaultConverter = func(val reflect.Value, _ reflect.Type) reflect.Value {
		return val
	}
	defaultConverter = func(val reflect.Value, typ reflect.Type) reflect.Value {
		return val.Convert(typ)
	}
)

// ConvertShallow converts value into specified type but deepest elements are not conversion target
func ConvertShallow(v interface{}, t reflect.Type) (reflect.Value, errors.Error) {
	return newConverter(v, shallowDefaultConverter).convert(t)
}

// Convert converts value into specified type, v as t
func Convert(v interface{}, t reflect.Type) (reflect.Value, errors.Error) {
	return newConverter(v, defaultConverter).convert(t)
}

type (
	converter struct {
		v                interface{}
		defaultConverter func(reflect.Value, reflect.Type) reflect.Value
	}
)

func newConverter(v interface{}, defaultConverter func(reflect.Value, reflect.Type) reflect.Value) *converter {
	return &converter{
		v:                v,
		defaultConverter: defaultConverter,
	}
}

func (s *converter) newConverter(v interface{}) *converter {
	return newConverter(v, s.defaultConverter)
}

func (s *converter) valueOf() reflect.Value {
	return reflect.ValueOf(s.v)
}

func (s *converter) typeOf() reflect.Type {
	return reflect.TypeOf(s.v)
}

func (s *converter) convert(t reflect.Type) (ret reflect.Value, e errors.Error) {
	defer func() {
		if err := recover(); err != nil {
			ret = reflect.Zero(t)
			e = errors.NewError().SetCode(errors.Validate).SetError(fmt.Errorf("invalid conversion: %v", err))
		}
	}()

	switch t.Kind() {
	case reflect.Array:
		return s.convertArray(t)
	case reflect.Chan:
		return s.convertChan(t)
	case reflect.Map:
		return s.convertMap(t)
	case reflect.Slice:
		return s.convertSlice(t)
	default:
		if s.defaultConverter != nil {
			return s.defaultConverter(s.valueOf(), t), nil
		}
		return s.valueOf(), nil
	}
}

func (s *converter) convertChan(t reflect.Type) (reflect.Value, errors.Error) {
	return reflect.MakeChan(t, s.valueOf().Cap()), nil
}

func (s *converter) convertArray(t reflect.Type) (reflect.Value, errors.Error) {
	sv := s.valueOf()
	arrayPtr := reflect.New(t)
	for i := 0; i < arrayPtr.Len(); i++ {
		x := sv.Index(i)
		y, err := s.newConverter(x.Interface()).convert(t.Elem())
		if err != nil {
			return reflect.Zero(t), err
		}
		arrayPtr.Index(i).Set(y)
	}
	return arrayPtr, nil
}

func (s *converter) convertSlice(t reflect.Type) (reflect.Value, errors.Error) {
	sv := s.valueOf()
	slicePtr := reflect.MakeSlice(t, sv.Len(), sv.Len())
	for i := 0; i < sv.Len(); i++ {
		x := sv.Index(i)
		y, err := s.newConverter(x.Interface()).convert(t.Elem())
		if err != nil {
			return reflect.Zero(t), err
		}
		slicePtr.Index(i).Set(y)
	}
	return slicePtr, nil
}

func (s *converter) convertMap(t reflect.Type) (reflect.Value, errors.Error) {
	sv := s.valueOf()
	mapPtr := reflect.MakeMapWithSize(t, sv.Len())
	mapIter := sv.MapRange()
	for mapIter.Next() {
		k, v := mapIter.Key(), mapIter.Value()
		ck, keyErr := s.newConverter(k.Interface()).convert(t.Key())
		cv, valueErr := s.newConverter(v.Interface()).convert(t.Elem())
		if keyErr != nil {
			return reflect.Zero(t), keyErr
		}
		if valueErr != nil {
			return reflect.Zero(t), valueErr
		}
		mapPtr.SetMapIndex(ck, cv)
	}
	return mapPtr, nil
}
