package sorter

import (
	"sort"
	"tools/pkg/errors"
	"tools/pkg/functions/executor"
	"tools/pkg/functions/iterator"
)

type (
	// Executor is map executor
	Executor struct {
		hooks executor.Hookable
		f     Sorter
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

func NewExecutor(f Sorter, iter iterator.Iterator, options ...Option) (*Executor, errors.Error) {
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

func (s *Executor) Execute() (interface{}, error) {
	s.hooks.Execute(executor.BeforeHook)
	defer s.hooks.Execute(executor.AfterHook)
	slice, err := iterator.ToSlice(s.iter)
	if err != nil {
		return nil, err
	}
	var sError error
	sort.SliceStable(slice, func(i, j int) bool {
		s.hooks.Execute(executor.RunningHook)
		ret, err := s.f.Apply(slice[i], slice[j])
		if err != nil && sError == nil {
			sError = err
		}
		return ret
	})
	if sError != nil {
		return nil, sError
	}
	return slice, nil
}
