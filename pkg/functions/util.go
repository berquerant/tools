package functions

import (
	"tools/pkg/functions/filter"
	"tools/pkg/functions/fold"
	"tools/pkg/functions/mapper"
)

type (
	// Script can be converted into a stream
	Script interface {
		// Type is type of stream
		Type() ScriptType
		// Instance is some function
		Instance() interface{}
		// Option is stream option
		Option(int) interface{}
		NumOption() int
	}

	script struct {
		T ScriptType
		I interface{}
		O []interface{}
	}

	// ScriptBuilder builds stream from scripts
	ScriptBuilder interface {
		// Type set type of script
		Type(ScriptType) ScriptBuilder
		// Instance set instance of script
		Instance(interface{}) ScriptBuilder
		// Option append option of script
		Option(interface{}) ScriptBuilder
		Build() Script
	}

	scriptBuilder struct {
		sc *script
	}
)

func (s *script) Type() ScriptType         { return s.T }
func (s *script) Instance() interface{}    { return s.I }
func (s *script) Option(i int) interface{} { return s.O[i] }
func (s *script) NumOption() int           { return len(s.O) }

func NewScriptBuilder() ScriptBuilder {
	return &scriptBuilder{
		sc: &script{
			O: []interface{}{},
		},
	}
}

func (s *scriptBuilder) Type(st ScriptType) ScriptBuilder {
	s.sc.T = st
	return s
}

func (s *scriptBuilder) Instance(f interface{}) ScriptBuilder {
	s.sc.I = f
	return s
}

func (s *scriptBuilder) Option(v interface{}) ScriptBuilder {
	s.sc.O = append(s.sc.O, v)
	return s
}

func (s *scriptBuilder) Build() Script {
	return s.sc
}

func NewStreamFromScripts(in Stream, ss []Script) Stream {
	s := NewStreamBuilder(in)
	for _, x := range ss {
		s.Append(x)
	}
	return s.Build()
}

//go:generate stringer -type=ScriptType -output generated.scripttype_string.go
type ScriptType int

const (
	UnknownScriptType ScriptType = iota
	// MapScriptType for Map
	MapScriptType
	// FilterScriptType for Filter
	FilterScriptType
	// FoldScriptType for Fold
	FoldScriptType
	// SortScriptType for Sort
	SortScriptType
	// FlatScriptType for Flat
	FlatScriptType
	// LiftScriptType for Lift
	LiftScriptType
)

type (
	// StreamBuilder builds stream from scripts
	StreamBuilder interface {
		Append(Script) StreamBuilder
		Build() Stream
	}

	streamBuilder struct {
		st Stream
	}
)

func NewStreamBuilder(in Stream) StreamBuilder {
	return &streamBuilder{
		st: in,
	}
}

func (s *streamBuilder) Append(x Script) StreamBuilder {
	s.st = func() func(Script) Stream {
		switch x.Type() {
		case MapScriptType:
			return s.appendMap
		case FilterScriptType:
			return s.appendFilter
		case FoldScriptType:
			return s.appendFold
		case SortScriptType:
			return s.appendSort
		case FlatScriptType:
			return s.appendFlat
		case LiftScriptType:
			return s.appendLift
		}
		return func(Script) Stream { return s.st }
	}()(x)
	return s
}

func (s *streamBuilder) appendMap(x Script) Stream {
	opts := []mapper.Option{}
	for i := 0; i < x.NumOption(); i++ {
		if p, ok := x.Option(i).(mapper.Option); ok {
			opts = append(opts, p)
		}
	}
	return s.st.Map(x.Instance(), opts...)
}

func (s *streamBuilder) appendFilter(x Script) Stream {
	opts := []filter.Option{}
	for i := 0; i < x.NumOption(); i++ {
		if p, ok := x.Option(i).(filter.Option); ok {
			opts = append(opts, p)
		}
	}
	return s.st.Filter(x.Instance(), opts...)
}

func (s *streamBuilder) appendFold(x Script) Stream {
	opts := []fold.Option{}
	for i := 0; i < x.NumOption(); i++ {
		if p, ok := x.Option(i).(fold.Option); ok {
			opts = append(opts, p)
		}
	}
	return s.st.Fold(x.Instance(), opts...)
}

func (s *streamBuilder) appendSort(x Script) Stream {
	return s.st.Sort(x.Instance())
}

func (s *streamBuilder) appendFlat(x Script) Stream {
	return s.st.Flat()
}

func (s *streamBuilder) appendLift(x Script) Stream {
	return s.st.Lift()
}

func (s *streamBuilder) Build() Stream {
	return s.st
}
