package executor

type (
	Hook func()

	Hookable interface {
		AddHook(HookType, Hook) Hookable
		GetHook(HookType) []Hook
		// Execute executes hooks of the hook type
		Execute(HookType)
	}

	hookable struct {
		hooks map[HookType][]Hook
	}
)

//go:generate stringer -type=HookType -output generated.hooktype_string.go
type HookType int

const (
	UnknownHook HookType = iota
	BeforeHook
	AfterHook
	RunningHook
)

func NewHookable() Hookable {
	return &hookable{
		hooks: map[HookType][]Hook{},
	}
}

func (s *hookable) AddHook(ht HookType, h Hook) Hookable {
	d, ok := s.hooks[ht]
	if !ok {
		d = []Hook{}
	}
	s.hooks[ht] = append(d, h)
	return s
}

func (s *hookable) GetHook(ht HookType) []Hook {
	return s.hooks[ht]
}

func (s *hookable) Execute(ht HookType) {
	for _, h := range s.GetHook(ht) {
		h()
	}
}
