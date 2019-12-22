package flat

import (
	"tools/pkg/collections/stack"
	"tools/pkg/errors"
	"tools/pkg/functions/executor"
	"tools/pkg/functions/iterator"
)

type (
	// Executor is flat executor
	Executor struct {
		hooks executor.Hookable
		iter  iterator.Iterator
		ft    Type
	}
	// Option changes option of Executor
	Option func(*Executor)
)

//go:generate stringer -type=Type -output generated.type_string.go
type Type int

const (
	TypeUnknown Type = iota
	// TypeSimple for concat
	TypeSimple
	// TypePerfect for recursive concat
	TypePerfect
)

// WithType specifies flat function type
func WithType(ft Type) Option {
	return func(s *Executor) {
		s.ft = ft
	}
}

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
		ft:    TypeSimple,
	}
	for _, opt := range options {
		opt(executor)
	}
	return executor, nil
}

func (s *Executor) Execute() iterator.Iterator {
	switch s.ft {
	case TypePerfect:
		return s.executePerfect()
	case TypeSimple:
		return s.executeSimple()
	}
	return nil
}

// executePerfect flat an iterator recursively.
// this yields elements that cannot be an iterator
func (s *Executor) executePerfect() iterator.Iterator {
	s.hooks.Execute(executor.BeforeHook, s.iter)
	var (
		stk   = stack.New()
		iFunc func() (interface{}, error)
	)
	stk.Push(s.iter)
	iFunc = func() (interface{}, error) {
		top, err := stk.Peep()
		if err != nil {
			s.hooks.Execute(executor.AfterHook)
			return nil, iterator.EOI
		}
		x, err := top.(iterator.Iterator).Next()
		if err == iterator.EOI {
			_, _ = stk.Pop()
			return iFunc()
		}
		if err != nil {
			return nil, err
		}
		if iterator.CanBeSingleValueIterator(x) {
			s.hooks.Execute(executor.RunningHook, x)
			return x, nil
		}
		stk.Push(iterator.MustNew(x))
		return iFunc()
	}
	return iterator.MustNew(iterator.Func(iFunc))
}

// executeSimple flat an iterator 1 level
func (s *Executor) executeSimple() iterator.Iterator {
	s.hooks.Execute(executor.BeforeHook, s.iter)
	var (
		top   iterator.Iterator
		iFunc func() (interface{}, error)
	)
	iFunc = func() (interface{}, error) {
		if top == nil {
			elem, err := s.iter.Next()
			if err != nil {
				if err == iterator.EOI {
					s.hooks.Execute(executor.AfterHook)
				}
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
		s.hooks.Execute(executor.RunningHook, x)
		return x, nil
	}
	return iterator.MustNew(iterator.Func(iFunc))
}
