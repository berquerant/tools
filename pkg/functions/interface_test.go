package functions_test

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"
	"tools/pkg/functions"
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
			Tester: functions.IsAggregator,
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
		return st.Fold(s.F, s.IV), nil
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

func (s *funcTuplesBuilder) Build() []*funcTuple {
	return s.v
}

func (s *streamTestcase) Test(t *testing.T) {
	st := functions.NewStream(iterator.MustNew(s.Data))
	for i, ft := range s.Translators {
		ret, err := ft.apply(st)
		if err != nil {
			t.Errorf("%d th translator returns error: %v", i, err)
			return
		}
		st = ret
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
	}

	for _, tt := range testcases {
		t.Run(tt.Comment, func(t *testing.T) {
			tt.Test(t)
		})
	}
}