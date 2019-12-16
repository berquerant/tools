package consume

import (
	"tools/pkg/errors"
	"tools/pkg/functions/executor"
	"tools/pkg/functions/iterator"
)

type (
	// Executor is consume executor
	Executor struct {
		hooks executor.Hookable
		f     Consumer
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

func NewExecutor(f Consumer, iter iterator.Iterator, options ...Option) (*Executor, errors.Error) {
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

func (s *Executor) Execute() error {
	s.hooks.Execute(executor.BeforeHook)
	defer s.hooks.Execute(executor.AfterHook)
	for {
		x, err := s.iter.Next()
		if err == iterator.EOI {
			return nil
		}
		if err != nil {
			return err
		}
		s.hooks.Execute(executor.RunningHook)
		if err := s.f.Apply(x); err != nil {
			return err
		}
	}
}
