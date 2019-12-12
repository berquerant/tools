package functions_test

import (
	"fmt"
	"testing"
	"tools/pkg/functions"
	"tools/pkg/functions/iterator"
)

type kv struct {
	k, v interface{}
}

func TestStream(t *testing.T) {
	err := functions.NewExtendedStream(
		functions.NewStream(
			iterator.MustNew([]*kv{
				{k: "alice", v: 10},
				{k: "bob", v: 18},
				{k: "catherine", v: 20},
				{k: "dan", v: 40},
				{k: "ellis", v: 16},
				{k: "fred", v: 19},
			}),
		),
	).Filter(
		func(v *kv) bool { return v.v.(int) < 20 },
	).Sort(
		func(x, y *kv) bool { return x.v.(int) < y.v.(int) },
	).Map(
		func(v *kv) string { return fmt.Sprintf("%v.%v", v.v, v.k) },
	).Fold(
		func(x string, acc int) int { return len(x) + acc }, 0,
	).Consume(func(v int) {
		t.Logf("got %d", v)
	})
	if err != nil {
		t.Error(err)
	}
}
