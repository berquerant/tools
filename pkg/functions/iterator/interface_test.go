package iterator_test

import (
	"testing"
	"tools/pkg/functions/iterator"

	"github.com/google/go-cmp/cmp"
)

type (
	iteratorTestcase struct {
		Comment string
		Data    interface{}
		Result  []interface{}
	}
)

func (s *iteratorTestcase) Test(t *testing.T) {
	it, err := iterator.New(s.Data)
	if err != nil {
		t.Error(err)
	}
	ret, err := iterator.ToSlice(it)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(ret, s.Result) {
		t.Errorf("  actual: %v\nexpected: %v", ret, s.Result)
	}
}

func TestIterator(t *testing.T) {
	testcases := []*iteratorTestcase{
		&iteratorTestcase{
			Comment: "nil iterator",
			Data:    nil,
			Result:  []interface{}{},
		},
		&iteratorTestcase{
			Comment: "single string",
			Data:    "answer",
			Result:  []interface{}{"answer"},
		},
		&iteratorTestcase{
			Comment: "iterator",
			Data:    iterator.MustNew("iter"),
			Result:  []interface{}{"iter"},
		},
		&iteratorTestcase{
			Comment: "array",
			Data:    [3]int{1, 2, 3},
			Result:  []interface{}{1, 2, 3},
		},
		&iteratorTestcase{
			Comment: "slice",
			Data:    []string{"a", "b", "c"},
			Result:  []interface{}{"a", "b", "c"},
		},
		&iteratorTestcase{
			Comment: "channel",
			Data: func() chan string {
				c := make(chan string, 3)
				defer close(c)
				c <- "no"
				c <- "where"
				c <- "man"
				return c
			}(),
			Result: []interface{}{"no", "where", "man"},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.Comment, func(t *testing.T) {
			tt.Test(t)
		})
	}
}
