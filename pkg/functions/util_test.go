package functions_test

import (
	"fmt"
	"strings"
	"testing"
	"tools/pkg/functions"
	"tools/pkg/functions/executor"
	"tools/pkg/functions/fold"
	"tools/pkg/functions/iterator"

	"github.com/google/go-cmp/cmp"
)

type (
	row struct {
		T functions.ScriptType
		I interface{}
		O []interface{}
	}
)

func (s *row) Script() functions.Script {
	sb := functions.NewScriptBuilder()
	sb.Type(s.T)
	if s.I != nil {
		sb.Instance(s.I)
	}
	for _, x := range s.O {
		sb.Option(x)
	}
	return sb.Build()
}

type (
	streamBuilderTestcase struct {
		Comment string
		Rows    []row
		Data    interface{}
		Result  []interface{}
	}
)

func (s *streamBuilderTestcase) Test(t *testing.T) {
	sb := functions.NewStreamBuilder(functions.NewStream(iterator.MustNew(s.Data)))
	for _, r := range s.Rows {
		sb.Append(r.Script())
	}
	st := sb.Build()
	actual, err := iterator.ToSlice(st)
	if err != nil {
		t.Errorf("to slice: %v", err)
	}
	if err := st.Err(); err != nil {
		t.Errorf("stream error: %v", err)
	}
	if !cmp.Equal(actual, s.Result) {
		t.Errorf("not expected result:\n  actual(%#v)\nexpected(%#v)", actual, s.Result)
	}
}

func TestStreamBuilder(t *testing.T) {
	testcases := []*streamBuilderTestcase{
		&streamBuilderTestcase{
			// awk '{d[toupper($0)]++} END {for (k in d) if (d[k] > 1) print k, d[k]}' | sort -nk 2 with handicaps
			Comment: "stat-with-handicaps",
			Data: []string{
				"Romania",
				"Norway",
				"Romania",
				"England",
				"China",
				"Greece",
				"Canada",
				"Slovenia",
				"Nepal",
				"Romania",
				"Slovakia",
				"Iran",
				"Greece",
				"Nepal",
				"Slovakia",
				"Turkey",
				"Romania",
				"Turkey",
				"United States",
				"Tunisia",
				"New Zealand",
				"Romania",
				"Turkey",
				"Nepal",
				"Kyrgyz Republic",
				"Romania",
				"Slovakia",
				"Turkey",
				"England",
				"United States",
			},
			Result: []interface{}{
				"UNITED STATES 2",
				"GREECE 2",
				"NEPAL 3",
				"TURKEY 4",
				"SLOVAKIA 5",
				"ROMANIA 6",
				"ENGLAND 12",
			},
			Rows: []row{
				{
					T: functions.MapScriptType,
					I: func(x string) string {
						return strings.ToUpper(x)
					},
				},
				{
					T: functions.FoldScriptType,
					I: func(x string, d map[string]int) map[string]int {
						d[x]++
						return d
					},
					O: []interface{}{
						fold.WithType(fold.TypeR),
						fold.WithInitialValue(map[string]int{
							"ENGLAND":  10,
							"SLOVAKIA": 2,
						}),
						fold.WithHook(executor.BeforeHook, func() {
							t.Log("before fold")
						}),
						fold.WithHook(executor.AfterHook, func() {
							t.Log("after fold")
						}),
					},
				},
				{
					T: functions.FlatScriptType,
				},
				{
					T: functions.FilterScriptType,
					I: func(x iterator.KV) bool {
						return x.V().(int) > 1
					},
				},
				{
					T: functions.SortScriptType,
					I: func(x, y iterator.KV) bool {
						return x.K().(string) > y.K().(string)
					},
				},
				{
					T: functions.SortScriptType,
					I: func(x, y iterator.KV) bool {
						return x.V().(int) < y.V().(int)
					},
				},
				{
					T: functions.MapScriptType,
					I: func(x iterator.KV) string {
						return fmt.Sprintf("%v %v", x.K(), x.V())
					},
				},
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.Comment, func(t *testing.T) {
			tt.Test(t)
		})
	}
}
