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
func WithHook(ht executor.HookType, h interface{}) Option {
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
	s.hooks.Execute(executor.BeforeHook, s.iter)
	for {
		x, err := s.iter.Next()
		if err == iterator.EOI {
			s.hooks.Execute(executor.AfterHook)
			return nil
		}
		if err != nil {
			return err
		}
		s.hooks.Execute(executor.RunningHook, x)
		if err := s.f.Apply(x); err != nil {
			return err
		}
		s.hooks.Execute(executor.RunningResultHook)
	}
}
