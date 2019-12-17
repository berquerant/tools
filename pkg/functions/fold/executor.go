package fold

import (
	"fmt"
	"tools/pkg/errors"
	"tools/pkg/functions/executor"
	"tools/pkg/functions/iterator"
)

var (
	InvalidType = errors.NewError().SetCode(errors.Fold).SetError(fmt.Errorf("invalid fold type"))
)

type (
	Func func(f Aggregator, acc interface{}, iter iterator.Iterator) (interface{}, error)

	// Executor is fold executor
	Executor struct {
		hooks executor.Hookable
		agg   Aggregator
		iter  iterator.Iterator
		ft    Type
		iv    interface{}
	}

	// Option changes option of Executor
	Option func(*Executor)
)

//go:generate stringer -type=Type -output generated.type_string.go
type Type int

const (
	TypeUnknown Type = iota
	// TypeR for Foldr
	TypeR
	// TypeL for Foldl
	TypeL
	// TypeT for Foldt
	TypeT
	// TypeI for Foldi
	TypeI
)

var (
	validExecutorMap = map[Type]func(AggregatorType) bool{
		TypeR: func(at AggregatorType) bool { return at == RightAggregator || at == PerfectAggregator },
		TypeL: func(at AggregatorType) bool { return at == LeftAggregator || at == PerfectAggregator },
		TypeT: func(at AggregatorType) bool { return at == PerfectAggregator },
		TypeI: func(at AggregatorType) bool { return at == PerfectAggregator },
	}
	funcMap = map[Type]Func{
		TypeR: Foldr,
		TypeL: Foldl,
		TypeT: Foldt,
		TypeI: Foldi,
	}
)

func isValidExecutor(ft Type, at AggregatorType) bool {
	t, ok := validExecutorMap[ft]
	return ok && t(at)
}

// WithType specifies fold function type
func WithType(ft Type) Option {
	return func(s *Executor) {
		s.ft = ft
	}
}

// WithInitialValue specifies fold initial value
func WithInitialValue(v interface{}) Option {
	return func(s *Executor) {
		s.iv = v
	}
}

// WithHook add hook
func WithHook(ht executor.HookType, h executor.Hook) Option {
	return func(s *Executor) {
		s.hooks.AddHook(ht, h)
	}
}

// NewExector creates Executor with default fold type R and initial zero value
func NewExecutor(f Aggregator, iter iterator.Iterator, options ...Option) (*Executor, errors.Error) {
	executor := &Executor{
		hooks: executor.NewHookable(),
		agg:   f,
		iter:  iter,
		ft:    TypeR,
		iv:    f.IV(),
	}
	for _, opt := range options {
		opt(executor)
	}
	if !isValidExecutor(executor.ft, f.Type()) {
		return nil, InvalidType
	}
	return executor, nil
}

func (s *Executor) Execute() (interface{}, error) {
	s.hooks.Execute(executor.BeforeHook)
	defer s.hooks.Execute(executor.AfterHook)
	if f, ok := funcMap[s.ft]; ok {
		s.hooks.Execute(executor.RunningHook)
		return f(s.agg, s.iv, s.iter)
	}
	return nil, InvalidType
}

// Foldr requires aggregator :: a -> b -> b
func Foldr(f Aggregator, acc interface{}, iter iterator.Iterator) (interface{}, error) {
	x, err := iter.Next()
	if err == iterator.EOI {
		return acc, nil
	}
	if err != nil {
		return nil, err
	}
	ret, err := Foldr(f, acc, iter)
	if err != nil {
		return nil, err
	}
	return f.Apply(x, ret)
}

// Foldl requires aggregator :: b -> a -> b
func Foldl(f Aggregator, acc interface{}, iter iterator.Iterator) (interface{}, error) {
	x, err := iter.Next()
	if err == iterator.EOI {
		return acc, nil
	}
	if err != nil {
		return nil, err
	}
	ret, err := f.Apply(acc, x)
	if err != nil {
		return nil, err
	}
	return Foldl(f, ret, iter)
}

// Foldt requires aggregator :: a -> a -> a
func Foldt(f Aggregator, acc interface{}, iter iterator.Iterator) (interface{}, error) {
	x, err := iter.Next()
	if err == iterator.EOI {
		return acc, nil
	}
	if err != nil {
		return nil, err
	}
	y, err := iter.Next()
	if err == iterator.EOI {
		return x, nil
	}
	if err != nil {
		return nil, err
	}
	// unyield one
	var (
		isYieldedOne bool
		piter        = pairs(f, iterator.MustNew(iterator.Func(func() (interface{}, error) {
			if !isYieldedOne {
				isYieldedOne = true
				return y, nil
			}
			return iter.Next()
		})))
	)
	return Foldt(f, acc, piter)
}

// Foldi requires aggregator :: a -> a -> a
func Foldi(f Aggregator, acc interface{}, iter iterator.Iterator) (interface{}, error) {
	x, err := iter.Next()
	if err == iterator.EOI {
		return acc, nil
	}
	if err != nil {
		return nil, err
	}
	ret, err := Foldi(f, acc, pairs(f, iter))
	if err != nil {
		return nil, err
	}
	return f.Apply(x, ret)
}

func pairs(f Aggregator, iter iterator.Iterator) iterator.Iterator {
	var isEOI bool
	return iterator.MustNew(iterator.Func(func() (interface{}, error) {
		if isEOI {
			return nil, iterator.EOI
		}
		x, err := iter.Next()
		if err != nil {
			return nil, err
		}
		y, err := iter.Next()
		if err == iterator.EOI {
			isEOI = true
			return x, nil
		}
		if err != nil {
			return nil, err
		}
		return f.Apply(x, y)
	}))
}
