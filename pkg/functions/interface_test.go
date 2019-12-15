package functions_test

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"
	"tools/pkg/functions"
	"tools/pkg/functions/fold"
	"tools/pkg/functions/iterator"

	"github.com/google/go-cmp/cmp"
)

type (
	Person struct {
		Name    string `json:"name"`
		Surname string `json:"surname"`
		Gender  string `json:"gender"`
		Region  string `json:"region"`
	}
)

const (
	peopleJSON = `[{"name":"Stela","surname":"Țîrle","gender":"female","region":"Romania"},{"name":"Aud","surname":"Rønning","gender":"female","region":"Norway"},{"name":"Natalia","surname":"Mironescu","gender":"female","region":"Romania"},{"name":"Hannah","surname":"Hunt","gender":"female","region":"England"},{"name":"余","surname":"河","gender":"male","region":"China"},{"name":"Μένων","surname":"Βλαστός","gender":"male","region":"Greece"},{"name":"Declan","surname":"Patel","gender":"male","region":"Canada"},{"name":"Ana","surname":"Novak","gender":"female","region":"Slovenia"},{"name":"Gaurav","surname":"Regmi","gender":"male","region":"Nepal"},{"name":"Tincuța","surname":"Gliga","gender":"female","region":"Romania"},{"name":"Dáša","surname":"Repiský","gender":"female","region":"Slovakia"},{"name":"مائده","surname":"حمیدی","gender":"female","region":"Iran"},{"name":"Δαναός","surname":"Ακρίδας","gender":"male","region":"Greece"},{"name":"Khemkala","surname":"Nagarkoti","gender":"female","region":"Nepal"},{"name":"Kristína","surname":"Murgaš","gender":"female","region":"Slovakia"},{"name":"İsa","surname":"Baştürk","gender":"male","region":"Turkey"},{"name":"Florea","surname":"Averescu","gender":"male","region":"Romania"},{"name":"Aslı","surname":"Paşa","gender":"female","region":"Turkey"},{"name":"Raymond","surname":"Daniels","gender":"male","region":"United States"},{"name":"وريدة","surname":"العيّاري","gender":"female","region":"Tunisia"},{"name":"Mya","surname":"Ackroyd","gender":"female","region":"New Zealand"},{"name":"Lucrețiu","surname":"Rusescu","gender":"male","region":"Romania"},{"name":"Nesrin","surname":"Mustafa","gender":"female","region":"Turkey"},{"name":"Bindu","surname":"Raut","gender":"female","region":"Nepal"},{"name":"Talas","surname":"Sultanov","gender":"male","region":"Kyrgyz Republic"},{"name":"Svetlana","surname":"Pătraș","gender":"female","region":"Romania"},{"name":"Ján","surname":"Svrbík","gender":"male","region":"Slovakia"},{"name":"Zeynep","surname":"Karadağ","gender":"female","region":"Turkey"},{"name":"Toby","surname":"Richardson","gender":"male","region":"England"},{"name":"Stephen","surname":"Hill","gender":"male","region":"United States"}]`
)

func people() []Person {
	var ps []Person
	_ = json.Unmarshal([]byte(peopleJSON), &ps)
	return ps
}

type (
	funcType int

	funcTuple struct {
		T  funcType
		F  interface{}
		IV interface{}
	}

	funcTypeClassifier struct {
		Tester func(interface{}) bool
		T      funcType
	}

	streamTestcase struct {
		Comment     string
		Data        interface{}
		Translators []*funcTuple
		Result      []interface{}
	}
)

const (
	ftUnknown funcType = iota
	ftMapper
	ftPredicate
	ftAggregator
	ftConsumer
	ftSorter
	ftFlat // dummy
	ftLift // dummy
)

var (
	errNotSupportedFuncType = fmt.Errorf("not supported func type")
	funcTypeClassifiers     = []*funcTypeClassifier{
		&funcTypeClassifier{ // override mapper :: a -> bool
			Tester: functions.IsPredicate,
			T:      ftPredicate,
		},
		&funcTypeClassifier{
			Tester: functions.IsMapper,
			T:      ftMapper,
		},
		&funcTypeClassifier{
			Tester: fold.IsAggregator,
			T:      ftAggregator,
		},
		&funcTypeClassifier{
			Tester: functions.IsConsumer,
			T:      ftConsumer,
		},
		&funcTypeClassifier{
			Tester: functions.IsSorter,
			T:      ftSorter,
		},
	}
)

func classifyFunc(f interface{}) funcType {
	for _, c := range funcTypeClassifiers {
		if c.Tester(f) {
			return c.T
		}
	}
	return ftUnknown
}

func (s *funcTuple) apply(st functions.Stream) (ret functions.Stream, e error) {
	defer func() {
		if err := recover(); err != nil {
			ret = nil
			e = fmt.Errorf("funcTuple.apply(): %v", err)
		}
	}()
	switch s.T {
	case ftMapper:
		return st.Map(s.F), nil
	case ftPredicate:
		return st.Filter(s.F), nil
	case ftAggregator:
		return st.Fold(s.F, fold.WithInitialValue(s.IV)), nil
	case ftSorter:
		return st.Sort(s.F), nil
	case ftFlat:
		return st.Flat(), nil
	case ftLift:
		return st.Lift(), nil
	default:
		return nil, errNotSupportedFuncType
	}
}

type (
	funcTuplesBuilder struct {
		v []*funcTuple
	}
)

func newFuncTuplesBuilder() *funcTuplesBuilder {
	return &funcTuplesBuilder{
		v: []*funcTuple{},
	}
}

func (s *funcTuplesBuilder) Append(f interface{}) *funcTuplesBuilder {
	s.v = append(s.v, &funcTuple{
		T: classifyFunc(f),
		F: f,
	})
	return s
}

func (s *funcTuplesBuilder) AppendWith(f, iv interface{}) *funcTuplesBuilder {
	s.v = append(s.v, &funcTuple{
		T:  classifyFunc(f),
		F:  f,
		IV: iv,
	})
	return s
}

func (s *funcTuplesBuilder) AppendType(ft funcType) *funcTuplesBuilder {
	s.v = append(s.v, &funcTuple{
		T: ft,
	})
	return s
}

func (s *funcTuplesBuilder) Build() []*funcTuple {
	return s.v
}

func (s *streamTestcase) Test(t *testing.T) {
	st := functions.NewStream(iterator.MustNew(s.Data))
	for i, ft := range s.Translators {
		ret, err := ft.apply(st)
		if err != nil {
			t.Errorf("%d th translator returns error: %v", i+1, err)
			return
		}
		st = ret
	}
	if err := st.Err(); err != nil {
		t.Errorf("stream error: %v", err)
	}
	actual, err := iterator.ToSlice(st)
	if err != nil {
		t.Errorf("cannot get result: %v", err)
		return
	}
	if !cmp.Equal(actual, s.Result) {
		t.Errorf("not expected result:\n  actual(%#v)\nexpected(%#v)", actual, s.Result)
	}
}

func TestStream(t *testing.T) {
	testcases := []*streamTestcase{
		&streamTestcase{
			Comment: "mapper-1-simple",
			Data:    people()[0:1],
			Translators: newFuncTuplesBuilder().
				Append(func(x Person) string {
					return strings.ToLower(x.Region)
				}).Build(),
			Result: func() []interface{} {
				return []interface{}{
					strings.ToLower(people()[0].Region),
				}
			}(),
		},
		&streamTestcase{
			Comment: "mapper-all-simple",
			Data:    people(),
			Translators: newFuncTuplesBuilder().
				Append(func(x Person) string {
					return strings.ToLower(x.Region)
				}).Build(),
			Result: func() []interface{} {
				r := make([]interface{}, len(people()))
				for i, p := range people() {
					r[i] = strings.ToLower(p.Region)
				}
				return r
			}(),
		},
		&streamTestcase{
			Comment: "mapper-all-complex",
			Data:    people(),
			Translators: newFuncTuplesBuilder().
				Append(func(x Person) struct {
					S        string
					Instance Person
				} {
					return struct {
						S        string
						Instance Person
					}{
						S:        fmt.Sprintf("%s/%s/%s/%s", x.Name, x.Surname, x.Gender, x.Region),
						Instance: x,
					}
				}).Build(),
			Result: func() []interface{} {
				r := make([]interface{}, len(people()))
				for i, p := range people() {
					r[i] = struct {
						S        string
						Instance Person
					}{
						S:        fmt.Sprintf("%s/%s/%s/%s", p.Name, p.Surname, p.Gender, p.Region),
						Instance: p,
					}
				}
				return r
			}(),
		},
		&streamTestcase{
			Comment: "lift-no-content",
			Data:    nil,
			Translators: newFuncTuplesBuilder().
				AppendType(ftLift).Build(),
			Result: []interface{}{},
		},
		&streamTestcase{
			Comment: "lift-1",
			Data:    []int{1},
			Translators: newFuncTuplesBuilder().
				AppendType(ftLift).Build(),
			Result: []interface{}{[]int{1}},
		},
		&streamTestcase{
			Comment: "lift-people",
			Data:    people(),
			Translators: newFuncTuplesBuilder().
				AppendType(ftLift).Build(),
			Result: []interface{}{people()},
		},
		&streamTestcase{
			Comment: "lift-flat-people",
			Data:    people(),
			Translators: newFuncTuplesBuilder().
				AppendType(ftLift).
				AppendType(ftFlat).Build(),
			Result: func() []interface{} {
				ps := people()
				r := make([]interface{}, len(ps))
				for i, p := range ps {
					r[i] = p
				}
				return r
			}(),
		},

		&streamTestcase{
			Comment: "flat-no-content",
			Data:    nil,
			Translators: newFuncTuplesBuilder().
				AppendType(ftFlat).Build(),
			Result: []interface{}{},
		},
		&streamTestcase{
			Comment: "flat-ints-1",
			Data: [][]int{
				[]int{1},
			},
			Translators: newFuncTuplesBuilder().
				AppendType(ftFlat).Build(),
			Result: []interface{}{1},
		},
		&streamTestcase{
			Comment: "flat-strings",
			Data: [][]string{
				[]string{"f", "l", "a", "t"},
				[]string{"str", "ring", "s"},
			},
			Translators: newFuncTuplesBuilder().
				AppendType(ftFlat).Build(),
			Result: []interface{}{"f", "l", "a", "t", "str", "ring", "s"},
		},
		&streamTestcase{
			Comment: "flat-deep",
			Data: [][][]int{
				[][]int{
					[]int{1},
					[]int{2, 3},
				},
				[][]int{
					[]int{123},
				},
			},
			Translators: newFuncTuplesBuilder().
				AppendType(ftFlat).Build(),
			Result: []interface{}{[]int{1}, []int{2, 3}, []int{123}},
		},
		&streamTestcase{
			Comment: "flat-lift-strings",
			Data: [][]string{
				[]string{"fl", "at"},
				[]string{"l", "ift"},
			},
			Translators: newFuncTuplesBuilder().
				AppendType(ftFlat).
				AppendType(ftLift).Build(),
			Result: []interface{}{[]string{
				"fl", "at", "l", "ift",
			}},
		},
		&streamTestcase{
			Comment: "sort-no-content",
			Data:    nil,
			Translators: newFuncTuplesBuilder().
				Append(func(x, y int) bool {
					return x < y
				}).Build(),
			Result: []interface{}{},
		},
		&streamTestcase{
			Comment: "sort-int",
			Data:    []int{5, 4, 9, 2},
			Translators: newFuncTuplesBuilder().
				Append(func(x, y int) bool {
					return x < y
				}).Build(),
			Result: []interface{}{2, 4, 5, 9},
		},
		&streamTestcase{
			Comment: "sort-people",
			Data:    people(),
			Translators: newFuncTuplesBuilder().
				Append(func(x, y Person) bool {
					return fmt.Sprintf("%s-%s", x.Name, x.Surname) < fmt.Sprintf("%s-%s", y.Name, y.Surname)
				}).Build(),
			Result: func() []interface{} {
				ps := people()
				sort.SliceStable(ps, func(i, j int) bool {
					return fmt.Sprintf("%s-%s", ps[i].Name, ps[i].Surname) < fmt.Sprintf("%s-%s", ps[j].Name, ps[j].Surname)
				})
				ret := make([]interface{}, len(ps))
				for i, p := range ps {
					ret[i] = p
				}
				return ret
			}(),
		},
		&streamTestcase{
			Comment: "filter-no-result",
			Data:    people(),
			Translators: newFuncTuplesBuilder().
				Append(func(Person) bool {
					return false
				}).Build(),
			Result: []interface{}{},
		},
		&streamTestcase{
			Comment: "filter-no-filter",
			Data:    people(),
			Translators: newFuncTuplesBuilder().
				Append(func(Person) bool {
					return true
				}).Build(),
			Result: func() []interface{} {
				r := make([]interface{}, len(people()))
				for i, p := range people() {
					r[i] = p
				}
				return r
			}(),
		},
		&streamTestcase{
			Comment: "filter-filtered-result",
			Data:    people(),
			Translators: newFuncTuplesBuilder().
				Append(func(x Person) bool {
					return x.Gender == "male"
				}).Build(),
			Result: func() []interface{} {
				r := []interface{}{}
				for _, p := range people() {
					if p.Gender == "male" {
						r = append(r, p)
					}
				}
				return r
			}(),
		},
		&streamTestcase{
			Comment: "aggregate-simple",
			Data:    people(),
			Translators: newFuncTuplesBuilder().
				AppendWith(func(_ Person, acc int) int {
					return 1 + acc
				}, 0).Build(),
			Result: []interface{}{len(people())},
		},
		&streamTestcase{
			Comment: "aggregate-complex",
			Data:    people(),
			Translators: newFuncTuplesBuilder().
				AppendWith(func(p Person, acc []int) []int {
					if p.Gender == "male" {
						return acc
					}
					return append([]int{len(p.Name) + len(p.Surname)}, acc...)
				}, []int{}).Build(),
			Result: func() []interface{} {
				r := []int{}
				for _, p := range people() {
					if p.Gender == "male" {
						continue
					}
					r = append(r, len(p.Name)+len(p.Surname))
				}
				return []interface{}{r}
			}(),
		},
		&streamTestcase{
			Comment: "mix",
			Data:    people(),
			Translators: newFuncTuplesBuilder().
				Append(func(p Person) bool {
					return p.Gender == "female"
				}).Append(func(p Person) string {
				return p.Region
			}).AppendWith(func(x string, acc map[string]int) map[string]int {
				acc[x]++
				return acc
			}, map[string]int{}).Append(func(x map[string]int) []string {
				ret := []string{}
				_ = functions.NewStream(iterator.MustNew(x)).Map(func(x iterator.KV) string {
					return fmt.Sprintf("%v/%v", x.K(), x.V())
				}).Consume(func(x string) {
					ret = append(ret, x)
				})
				sort.Strings(ret)
				return ret
			}).Build(),
			Result: func() []interface{} {
				regionCount := map[string]int{}
				for _, p := range people() {
					if p.Gender != "female" {
						continue
					}
					regionCount[p.Region]++
				}
				sortedKeys := make([]string, len(regionCount))
				var i int
				for k := range regionCount {
					sortedKeys[i] = k
					i++
				}
				sort.Strings(sortedKeys)
				ret := make([]string, len(sortedKeys))
				for i, k := range sortedKeys {
					ret[i] = fmt.Sprintf("%s/%d", k, regionCount[k])
				}
				return []interface{}{ret}
			}(),
		},
		&streamTestcase{
			Comment: "mix2",
			Data:    people(),
			Translators: newFuncTuplesBuilder().
				Append(func(p Person) string {
					return strings.ToUpper(p.Region)
				}).AppendWith(func(x string, d map[string]int) map[string]int {
				d[x]++
				return d
			}, map[string]int{},
			).AppendType(
				ftFlat,
			).Append(func(x iterator.KV) bool {
				return x.V().(int) > 1
			}).Append(func(x, y iterator.KV) bool {
				return x.K().(string) < y.K().(string)
			}).Append(func(x, y iterator.KV) bool {
				return x.V().(int) < y.V().(int)
			}).Append(func(x iterator.KV) string {
				return fmt.Sprintf("%v:%v", x.K(), x.V())
			}).Build(),
			Result: func() []interface{} {
				d := map[string]int{}
				for _, p := range people() {
					d[strings.ToUpper(p.Region)]++
				}
				type Cell struct {
					Region string
					N      int
				}
				cells := []*Cell{}
				for k, v := range d {
					if v <= 1 {
						continue
					}
					cells = append(cells, &Cell{
						Region: k,
						N:      v,
					})
				}
				sort.SliceStable(cells, func(i, j int) bool {
					return cells[i].Region < cells[j].Region
				})
				sort.SliceStable(cells, func(i, j int) bool {
					return cells[i].N < cells[j].N
				})
				r := make([]interface{}, len(cells))
				for i, c := range cells {
					r[i] = fmt.Sprintf("%s:%d", c.Region, c.N)
				}
				return r
			}(),
		},
	}

	for _, tt := range testcases {
		t.Run(tt.Comment, func(t *testing.T) {
			tt.Test(t)
		})
	}
}

type (
	streamConsumeTestcase struct {
		Comment string
		Data    []interface{}
	}
)

func (s *streamConsumeTestcase) Test(t *testing.T) {
	var i int
	if err := functions.NewStream(iterator.MustNew(s.Data)).Consume(func(x interface{}) {
		if !cmp.Equal(x, s.Data[i]) {
			t.Errorf("not equal\n  actual: %#v\nexpected: %#v", x, s.Data[i])
		}
		i++
	}); err != nil {
		t.Error(err)
	}
}

func TestStreamConsume(t *testing.T) {
	testcases := []*streamConsumeTestcase{
		&streamConsumeTestcase{
			Comment: "no-content",
			Data:    nil,
		},
		&streamConsumeTestcase{
			Comment: "int-1",
			Data:    []interface{}{1},
		},
		&streamConsumeTestcase{
			Comment: "int2-2",
			Data:    []interface{}{[]int{1}, []int{2}},
		},
		&streamConsumeTestcase{
			Comment: "string-3",
			Data:    []interface{}{"1", "2", "3"},
		},
		&streamConsumeTestcase{
			Comment: "people",
			Data: func() []interface{} {
				r := make([]interface{}, len(people()))
				for i, p := range people() {
					r[i] = p
				}
				return r
			}(),
		},
	}

	for _, tt := range testcases {
		t.Run(tt.Comment, func(t *testing.T) {
			tt.Test(t)
		})
	}
}

func TestStreamAs(t *testing.T) {
	convert := func(d interface{}) functions.Stream {
		return functions.NewStream(iterator.MustNew(d))
	}

	testcases := []struct {
		Comment string
		Test    func(*testing.T)
	}{
		{
			Comment: "no-content",
			Test: func(t *testing.T) {
				var (
					d = []int{}
					r []int
				)
				if err := convert(d).As(&r); err != nil {
					t.Error(err)
				}
				if len(r) != 0 {
					t.Errorf("got %#v", r)
				}
			},
		},
		{
			Comment: "int-1",
			Test: func(t *testing.T) {
				var (
					d = []int{1}
					r []int
				)
				if err := convert(d).As(&r); err != nil {
					t.Error(err)
				}
				if !cmp.Equal(r, d) {
					t.Errorf("got %#v", r)
				}
			},
		},
		{
			Comment: "string2-2",
			Test: func(t *testing.T) {
				var (
					d = [][]string{
						[]string{"1"},
						[]string{"2", "3"},
					}
					r [][]string
				)
				if err := convert(d).As(&r); err != nil {
					t.Error(err)
				}
				if !cmp.Equal(r, d) {
					t.Errorf("got %#v", r)
				}
			},
		},
		{
			Comment: "people",
			Test: func(t *testing.T) {
				var (
					d = people()
					r []Person
				)
				if err := convert(d).As(&r); err != nil {
					t.Error(err)
				}
				if !cmp.Equal(r, d) {
					t.Errorf("got %#v", r)
				}
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.Comment, tt.Test)
	}
}
