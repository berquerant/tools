package filter

import (
	"tools/pkg/errors"
	"tools/pkg/functions/executor"
	"tools/pkg/functions/iterator"
)

type (
	// Executor is filter executor
	Executor struct {
		hooks executor.Hookable
		f     Predicate
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

func NewExecutor(f Predicate, iter iterator.Iterator, options ...Option) (*Executor, errors.Error) {
	executor := &Executor{
		hooks: executor.NewHookable(),
		f:     f,
		iter:  iter,
	}
	for _, opt := range options {
		opt(executor)
	}
	return executor, nil
}

func (s *Executor) Execute() iterator.Iterator {
	s.hooks.Execute(executor.BeforeHook, s.iter)
	var (
		iFunc func() (interface{}, error)
	)
	iFunc = func() (interface{}, error) {
		x, err := s.iter.Next()
		if err != nil {
			if err == iterator.EOI {
				s.hooks.Execute(executor.AfterHook)
			}
			return nil, err
		}
		s.hooks.Execute(executor.RunningHook, x)
		ret, err := s.f.Apply(x)
		if err != nil {
			return nil, err
		}
		s.hooks.Execute(executor.RunningResultHook, ret)
		if !ret {
			return iFunc()
		}
		return x, nil
	}
	return iterator.MustNew(iterator.Func(iFunc))
}
