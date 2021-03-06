package functions_test

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"
	"tools/pkg/functions"
	"tools/pkg/functions/executor"
	"tools/pkg/functions/flat"
	"tools/pkg/functions/fold"
	"tools/pkg/functions/iterator"
	"tools/pkg/functions/mapper"

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
	streamTestcase struct {
		Comment string
		Data    interface{}
		Stream  func(functions.Stream) functions.Stream
		Result  []interface{}
	}
)

func (s *streamTestcase) Test(t *testing.T) {
	st := functions.NewStream(iterator.MustNew(s.Data))
	st = s.Stream(st)
	if err := st.Err(); err != nil {
		t.Errorf("stream error: %v", err)
	}
	actual, err := iterator.ToSlice(st)
	if err := st.Err(); err != nil {
		t.Errorf("stream error: %v", err)
	}
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
			Stream: func(s functions.Stream) functions.Stream {
				return s.Map(func(x Person) string {
					return strings.ToLower(x.Region)
				})
			},
			Result: func() []interface{} {
				return []interface{}{
					strings.ToLower(people()[0].Region),
				}
			}(),
		},
		&streamTestcase{
			Comment: "mapper-all-simple",
			Data:    people(),
			Stream: func(s functions.Stream) functions.Stream {
				return s.Map(func(x Person) string {
					return strings.ToLower(x.Region)
				})
			},
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
			Stream: func(s functions.Stream) functions.Stream {
				return s.Map(func(x Person) struct {
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
				})
			},
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
			Stream: func(s functions.Stream) functions.Stream {
				return s.Lift()
			},
			Result: []interface{}{},
		},
		&streamTestcase{
			Comment: "lift-1",
			Data:    []int{1},
			Stream: func(s functions.Stream) functions.Stream {
				return s.Lift()
			},
			Result: []interface{}{[]int{1}},
		},
		&streamTestcase{
			Comment: "lift-people",
			Data:    people(),
			Stream: func(s functions.Stream) functions.Stream {
				return s.Lift()
			},
			Result: []interface{}{people()},
		},
		&streamTestcase{
			Comment: "lift-flat-people",
			Data:    people(),
			Stream: func(s functions.Stream) functions.Stream {
				return s.Lift().Flat()
			},
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
			Stream: func(s functions.Stream) functions.Stream {
				return s.Flat()
			},
			Result: []interface{}{},
		},
		&streamTestcase{
			Comment: "flat-ints-1",
			Data: [][]int{
				[]int{1},
			},
			Stream: func(s functions.Stream) functions.Stream {
				return s.Flat()
			},
			Result: []interface{}{1},
		},
		&streamTestcase{
			Comment: "flat-strings",
			Data: [][]string{
				[]string{"f", "l", "a", "t"},
				[]string{"str", "ring", "s"},
			},
			Stream: func(s functions.Stream) functions.Stream {
				return s.Flat()
			},
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
			Stream: func(s functions.Stream) functions.Stream {
				return s.Flat()
			},
			Result: []interface{}{[]int{1}, []int{2, 3}, []int{123}},
		},
		&streamTestcase{
			Comment: "flat-lift-strings",
			Data: [][]string{
				[]string{"fl", "at"},
				[]string{"l", "ift"},
			},
			Stream: func(s functions.Stream) functions.Stream {
				return s.Flat().Lift()
			},
			Result: []interface{}{[]string{
				"fl", "at", "l", "ift",
			}},
		},
		&streamTestcase{
			Comment: "flat-perfect",
			Data: [][][]int{
				[][]int{
					[]int{
						1, 2,
					},
				},
				[][]int{
					[]int{3},
					[]int{4, 5, 6},
				},
			},
			Stream: func(s functions.Stream) functions.Stream {
				return s.Flat(flat.WithType(flat.TypePerfect))
			},
			Result: []interface{}{
				1, 2, 3, 4, 5, 6,
			},
		},
		&streamTestcase{
			Comment: "flat-perfect-interface",
			Data: []interface{}{
				1,
				[]int{2, 3},
				[][]int{
					[]int{4},
					[]int{5},
				},
				[][][]int{
					[][]int{
						[]int{6},
					},
				},
			},
			Stream: func(s functions.Stream) functions.Stream {
				return s.Flat(flat.WithType(flat.TypePerfect))
			},
			Result: []interface{}{
				1, 2, 3, 4, 5, 6,
			},
		},
		&streamTestcase{
			Comment: "sort-no-content",
			Data:    nil,
			Stream: func(s functions.Stream) functions.Stream {
				return s.Sort(func(x, y int) bool {
					return x < y
				})
			},
			Result: []interface{}{},
		},
		&streamTestcase{
			Comment: "sort-int",
			Data:    []int{5, 4, 9, 2},
			Stream: func(s functions.Stream) functions.Stream {
				return s.Sort(func(x, y int) bool {
					return x < y
				})
			},
			Result: []interface{}{2, 4, 5, 9},
		},
		&streamTestcase{
			Comment: "sort-people",
			Data:    people(),
			Stream: func(s functions.Stream) functions.Stream {
				return s.Sort(func(x, y Person) bool {
					return fmt.Sprintf("%s-%s", x.Name, x.Surname) < fmt.Sprintf("%s-%s", y.Name, y.Surname)
				})
			},
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
			Stream: func(s functions.Stream) functions.Stream {
				return s.Filter(func(Person) bool {
					return false
				})
			},
			Result: []interface{}{},
		},
		&streamTestcase{
			Comment: "filter-no-filter",
			Data:    people(),
			Stream: func(s functions.Stream) functions.Stream {
				return s.Filter(func(Person) bool {
					return true
				})
			},
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
			Stream: func(s functions.Stream) functions.Stream {
				return s.Filter(func(x Person) bool {
					return x.Gender == "male"
				})
			},
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
			Stream: func(s functions.Stream) functions.Stream {
				return s.Fold(func(_ Person, acc int) int {
					return 1 + acc
				})
			},
			Result: []interface{}{len(people())},
		},
		&streamTestcase{
			Comment: "aggregate-complex",
			Data:    people(),
			Stream: func(s functions.Stream) functions.Stream {
				return s.Fold(func(p Person, acc []int) []int {
					if p.Gender == "male" {
						return acc
					}
					return append([]int{len(p.Name) + len(p.Surname)}, acc...)
				}, fold.WithInitialValue([]int{}))
			},
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
			Stream: func(s functions.Stream) functions.Stream {
				return s.Filter(func(p Person) bool {
					return p.Gender == "female"
				}).Map(func(p Person) string {
					return p.Region
				}).Fold(func(x string, acc map[string]int) map[string]int {
					acc[x]++
					return acc
				}, fold.WithInitialValue(map[string]int{})).Map(func(x map[string]int) []string {
					ret := []string{}
					_ = functions.NewStream(iterator.MustNew(x)).Map(func(x iterator.KV) string {
						return fmt.Sprintf("%v/%v", x.K(), x.V())
					}).Consume(func(x string) {
						ret = append(ret, x)
					})
					sort.Strings(ret)
					return ret
				})
			},
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
			Stream: func(s functions.Stream) functions.Stream {
				return s.Map(func(p Person) string {
					return strings.ToUpper(p.Region)
				}).Fold(func(x string, d map[string]int) map[string]int {
					d[x]++
					return d
				}, fold.WithInitialValue(map[string]int{})).Flat().Filter(func(x iterator.KV) bool {
					return x.V().(int) > 1
				}).Sort(func(x, y iterator.KV) bool {
					return x.K().(string) < y.K().(string)
				}).Sort(func(x, y iterator.KV) bool {
					return x.V().(int) < y.V().(int)
				}).Map(func(x iterator.KV) string {
					return fmt.Sprintf("%v:%v", x.K(), x.V())
				})
			},
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

func TestStreamComplex(t *testing.T) {
	t.Run("map-hook", func(t *testing.T) {
		var (
			data = []int{
				1, 2, 3,
			}
			response = []int{
				101, 102, 103,
			}
			runningC       = make(chan int, len(data))
			runningResultC = make(chan int, len(data))
			resultC        = make(chan int, len(data))

			origin = functions.NewStream(iterator.MustNew(data)).Map(func(x int) int {
				return x + 100
			}, mapper.WithHook(executor.RunningHook, func(x int) {
				runningC <- x
			}), mapper.WithHook(executor.RunningResultHook, func(x int) {
				runningResultC <- x
			}))

			running       = functions.NewStream(iterator.MustNew(runningC))
			runningResult = functions.NewStream(iterator.MustNew(runningResultC))
			result        = functions.NewStream(iterator.MustNew(resultC))
		)
		if err := origin.Consume(func(x int) {
			resultC <- x
		}); err != nil {
			t.Error(err)
		}
		close(resultC)
		close(runningC)
		close(runningResultC)

		var (
			rSlice       = []int{}
			rResultSlice = []int{}
			resultSlice  = []int{}
		)
		if err := running.As(&rSlice); err != nil {
			t.Error(err)
		}
		if err := runningResult.As(&rResultSlice); err != nil {
			t.Error(err)
		}
		if err := result.As(&resultSlice); err != nil {
			t.Error(err)
		}

		if !cmp.Equal(rSlice, data) {
			t.Errorf("  actual: %v\nexpected: %v", rSlice, data)
		}
		if !cmp.Equal(resultSlice, response) {
			t.Errorf("  actual: %v\nexpected: %v", resultSlice, response)
		}
		if !cmp.Equal(rResultSlice, resultSlice) {
			t.Errorf("runningResult: %v\n       result: %v", rResultSlice, resultSlice)
		}
	})
}
