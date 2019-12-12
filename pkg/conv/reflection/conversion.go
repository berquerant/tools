package reflection

import (
	"fmt"
	"reflect"
	"tools/pkg/errors"
)

// Convert converts valut into specified type.
// v as t
func Convert(v, t interface{}) (reflect.Value, errors.Error) {
	return newConverter(v).convert(reflect.TypeOf(t))
}

type (
	converter struct {
		v interface{}
	}
)

func newConverter(v interface{}) *converter {
	return &converter{v: v}
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
	}
	return s.valueOf(), nil
}

func (s *converter) convertChan(t reflect.Type) (reflect.Value, errors.Error) {
	return reflect.MakeChan(t, s.valueOf().Cap()), nil
}

func (s *converter) convertArray(t reflect.Type) (reflect.Value, errors.Error) {
	sv := s.valueOf()
	arrayPtr := reflect.New(t)
	for i := 0; i < arrayPtr.Len(); i++ {
		x := sv.Index(i)
		y, err := newConverter(x.Interface()).convert(t.Elem())
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
		y, err := newConverter(x.Interface()).convert(t.Elem())
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
		ck, keyErr := newConverter(k.Interface()).convert(t.Key())
		cv, valueErr := newConverter(v.Interface()).convert(t.Elem())
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
