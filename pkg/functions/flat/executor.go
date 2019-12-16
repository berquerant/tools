package flat

import (
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
func WithHook(ht executor.HookType, h executor.Hook) Option {
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

func (s *Executor) Execute() iterator.Iterator {
	var (
		top      iterator.Iterator
		iFunc    func() (interface{}, error)
		isBefore = true
	)
	iFunc = func() (interface{}, error) {
		if top == nil {
			if isBefore {
				isBefore = false
				s.hooks.Execute(executor.BeforeHook)
			}
			elem, err := s.iter.Next()
			if err != nil {
				s.hooks.Execute(executor.AfterHook)
				return nil, err
			}
			top = iterator.MustNew(elem)
		}

		x, err := top.Next()
		if err == iterator.EOI {
			top = nil
			return iFunc()
		}
		if err != nil {
			return nil, err
		}
		s.hooks.Execute(executor.RunningHook)
		return x, nil
	}
	return iterator.MustNew(iterator.Func(iFunc))
}
