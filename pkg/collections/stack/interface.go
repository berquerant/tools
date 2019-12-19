package stack

import "fmt"

type (
	Stack interface {
		Push(interface{})
		Pop() (interface{}, error)
		Peep() (interface{}, error)
		Len() int
	}

	sliceStack struct {
		instance []interface{}
		index    int
	}
)

var (
	errEmpty = fmt.Errorf("empty")
)

func New() Stack {
	return &sliceStack{
		instance: []interface{}{},
		index:    -1,
	}
}

func (s *sliceStack) Len() int {
	return s.index + 1
}

func (s *sliceStack) Peep() (interface{}, error) {
	if s.index < 0 {
		return nil, errEmpty
	}
	return s.instance[s.index], nil
}

func (s *sliceStack) Push(v interface{}) {
	s.instance = append(s.instance, v)
	s.index++
}

func (s *sliceStack) Pop() (interface{}, error) {
	x, err := s.Peep()
	if err != nil {
		return nil, err
	}
	s.instance = s.instance[:s.index]
	s.index--
	return x, nil
}
