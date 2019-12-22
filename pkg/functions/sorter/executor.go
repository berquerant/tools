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
func WithHook(ht executor.HookType, h interface{}) Option {
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

func (s *Executor) Execute() (iterator.Iterator, error) {
	s.hooks.Execute(executor.BeforeHook, s.iter)
	slice, err := iterator.ToSlice(s.iter)
	if err != nil {
		return nil, err
	}
	var sError error
	sort.SliceStable(slice, func(i, j int) bool {
		s.hooks.Execute(executor.RunningHook, slice[i], slice[j])
		ret, err := s.f.Apply(slice[i], slice[j])
		if err != nil && sError == nil {
			sError = err
		}
		if err == nil {
			s.hooks.Execute(executor.RunningResultHook, ret)
		}
		return ret
	})
	if sError != nil {
		return nil, sError
	}
	defer s.hooks.Execute(executor.AfterHook)
	return iterator.MustNew(slice), nil
}
