package executor

import (
	"reflect"
	"tools/pkg/conv/reflection"
)

type (
	Hookable interface {
		// AddHook accepts any function
		AddHook(ht HookType, hook interface{}) Hookable
		GetHook(ht HookType) []interface{}
		// Execute executes hooks of the hook type.
		// executes functions that have appropriate size and types of arguments
		Execute(ht HookType, args ...interface{})
	}

	hookable struct {
		hooks map[HookType][]interface{}
	}
)

//go:generate stringer -type=HookType -output generated.hooktype_string.go
type HookType int

const (
	UnknownHook HookType = iota
	BeforeHook
	AfterHook
	RunningHook
	RunningResultHook
)

func NewHookable() Hookable {
	return &hookable{
		hooks: map[HookType][]interface{}{},
	}
}

func (s *hookable) AddHook(ht HookType, h interface{}) Hookable {
	if reflect.TypeOf(h).Kind() != reflect.Func {
		return s
	}
	d, ok := s.hooks[ht]
	if !ok {
		d = []interface{}{}
	}
	s.hooks[ht] = append(d, h)
	return s
}

func (s *hookable) GetHook(ht HookType) []interface{} {
	return s.hooks[ht]
}

func (s *hookable) Execute(ht HookType, args ...interface{}) {
	for _, h := range s.GetHook(ht) {
		s.execute(h, args...)
	}
}

func (s *hookable) execute(h interface{}, args ...interface{}) {
	t := reflect.TypeOf(h)
	if t.NumIn() != len(args) {
		return
	}
	vargs := make([]reflect.Value, len(args))
	for i, a := range args {
		v, err := reflection.ConvertShallow(a, t.In(i))
		if err != nil {
			return
		}
		vargs[i] = v
	}
	reflect.ValueOf(h).Call(vargs)
}
