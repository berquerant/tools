package functions

import (
	"fmt"
	"tools/pkg/errors"
	"tools/pkg/functions/iterator"
)

var (
	InvalidFoldType = errors.NewError().SetCode(errors.Fold).SetError(fmt.Errorf("invalid fold type"))
)

type (
	FoldFunc func(f Aggregator, acc interface{}, iter iterator.Iterator) (interface{}, error)

	// Folder is fold executor
	Folder struct {
		agg  Aggregator
		iter iterator.Iterator
		ft   FoldType
		iv   interface{}
	}

	// FoldOptionFunc changes option of Folder
	FoldOptionFunc func(*Folder)

	// FoldType indicates type of fold
	FoldType int
)

const (
	UnknownFold FoldType = iota
	// FoldTypeR for Foldr
	FoldTypeR
	// FoldTypeL for Foldl
	FoldTypeL
	// FoldTypeT for Foldt
	FoldTypeT
	// FoldTypeI for Foldi
	FoldTypeI
)

var (
	validFolderMap = map[FoldType]func(AggregatorType) bool{
		FoldTypeR: func(at AggregatorType) bool { return at == RightAggregator || at == PerfectAggregator },
		FoldTypeL: func(at AggregatorType) bool { return at == LeftAggregator || at == PerfectAggregator },
		FoldTypeT: func(at AggregatorType) bool { return at == PerfectAggregator },
		FoldTypeI: func(at AggregatorType) bool { return at == PerfectAggregator },
	}
	foldFuncMap = map[FoldType]FoldFunc{
		FoldTypeR: Foldr,
		FoldTypeL: Foldl,
		FoldTypeT: Foldt,
		FoldTypeI: Foldi,
	}
)

func isValidFolder(ft FoldType, at AggregatorType) bool {
	t, ok := validFolderMap[ft]
	return ok && t(at)
}

// WithFoldType specifies fold function type
func WithFoldType(ft FoldType) FoldOptionFunc {
	return func(s *Folder) {
		s.ft = ft
	}
}

// WithInitialValue specifies fold initial value
func WithInitialValue(v interface{}) FoldOptionFunc {
	return func(s *Folder) {
		s.iv = v
	}
}

// NewFolder creates Folder with default fold type R and initial zero value
func NewFolder(f Aggregator, iter iterator.Iterator, options ...FoldOptionFunc) (*Folder, errors.Error) {
	folder := &Folder{
		agg:  f,
		iter: iter,
		ft:   FoldTypeR,
		iv:   f.IV(),
	}
	for _, opt := range options {
		opt(folder)
	}
	if !isValidFolder(folder.ft, f.Type()) {
		return nil, InvalidFoldType
	}
	return folder, nil
}

func (s *Folder) Fold() (interface{}, error) {
	if f, ok := foldFuncMap[s.ft]; ok {
		return f(s.agg, s.iv, s.iter)
	}
	return nil, InvalidFoldType
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