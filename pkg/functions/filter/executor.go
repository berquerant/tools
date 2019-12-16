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
func WithHook(ht executor.HookType, h executor.Hook) Option {
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
	var (
		iFunc    func() (interface{}, error)
		isBefore = true
	)
	iFunc = func() (interface{}, error) {
		if isBefore {
			isBefore = false
			s.hooks.Execute(executor.BeforeHook)
		}
		x, err := s.iter.Next()
		if err != nil {
			s.hooks.Execute(executor.AfterHook)
			return nil, err
		}
		s.hooks.Execute(executor.RunningHook)
		ret, err := s.f.Apply(x)
		if err != nil {
			s.hooks.Execute(executor.AfterHook)
			return nil, err
		}
		if !ret {
			return iFunc()
		}
		return x, nil
	}
	return iterator.MustNew(iterator.Func(iFunc))
}
