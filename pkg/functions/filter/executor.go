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

func (s *Executor) Next() (interface{}, error) {
	s.hooks.Execute(executor.RunningHook)
	x, err := s.iter.Next()
	if err != nil {
		return nil, err
	}
	ret, err := s.f.Apply(x)
	if err != nil {
		return nil, err
	}
	if !ret {
		return s.Next()
	}
	return x, nil
}
