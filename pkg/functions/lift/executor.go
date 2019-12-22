package lift

import (
	"reflect"
	"tools/pkg/conv/reflection"
	"tools/pkg/errors"
	"tools/pkg/functions/executor"
	"tools/pkg/functions/iterator"
)

type (
	// Executor is flat executor
	Executor struct {
		hooks executor.Hookable
		iter  iterator.Iterator
	}
	// Option changes option of Executor
	Option func(*Executor)
)

// WithHook add hook
func WithHook(ht executor.HookType, h interface{}) Option {
	return func(s *Executor) {
		s.hooks.AddHook(ht, h)
	}
}

func NewExecutor(iter iterator.Iterator, options ...Option) (*Executor, errors.Error) {
	executor := &Executor{
		hooks: executor.NewHookable(),
		iter:  iter,
	}
	for _, opt := range options {
		opt(executor)
	}
	return executor, nil
}

func (s *Executor) Execute() (iterator.Iterator, error) {
	s.hooks.Execute(executor.BeforeHook, s.iter)
	var err error
	slice, err := iterator.ToSlice(s.iter)
	if err != nil {
		return nil, err
	}

	if len(slice) == 0 {
		return iterator.MustNew(iterator.Func(func() (interface{}, error) {
			defer s.hooks.Execute(executor.AfterHook)
			return nil, iterator.EOI
		})), nil
	}

	t := getCommonType(slice)
	newSlice, err := reflection.Convert(slice, reflect.SliceOf(t))
	if err != nil {
		return nil, err
	}
	var isBefore = true
	return iterator.MustNew(iterator.Func(func() (interface{}, error) {
		if isBefore {
			isBefore = false
			s.hooks.Execute(executor.RunningHook, s.iter)
			ret := newSlice.Interface()
			s.hooks.Execute(executor.RunningResultHook, ret)
			return ret, nil
		}
		defer s.hooks.Execute(executor.AfterHook)
		return nil, iterator.EOI
	})), nil
}

func getCommonType(v []interface{}) reflect.Type {
	defaultType := reflect.TypeOf([]interface{}{})
	if len(v) == 0 {
		return defaultType
	}
	t := reflect.TypeOf(v[0])
	for _, x := range v {
		if reflect.TypeOf(x).String() != t.String() {
			return defaultType
		}
	}
	return t
}
