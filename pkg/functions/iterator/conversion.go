package iterator

import "tools/pkg/errors"

type (
	// IE is a cell to contain element sent from the channel that made from iterator
	IE interface {
		I() interface{}
		E() error
	}
	ie struct {
		i interface{}
		e error
	}
)

func (s *ie) I() interface{} { return s.i }
func (s *ie) E() error       { return s.e }

// ToChan converts iterator into channel.
// The channel is closed when iterator reached the end or some error
func ToChan(iter Iterator) (<-chan IE, errors.Error) {
	ch := make(chan IE)
	go func() {
		for {
			x, err := iter.Next()
			if err == EOI {
				close(ch)
				return
			}
			if err != nil {
				ch <- &ie{e: err}
				close(ch)
				return
			}
			ch <- &ie{i: x}
		}
	}()
	return ch, nil
}

// ToSlice convertes iterator into slice
func ToSlice(iter Iterator) ([]interface{}, errors.Error) {
	ret := []interface{}{}
	for {
		x, err := iter.Next()
		if err == EOI {
			return ret, nil
		}
		if err != nil {
			return nil, errors.NewError().SetCode(errors.Iterator).SetError(err)
		}
		ret = append(ret, x)
	}
}

// ToFunc converts iterator into function
func ToFunc(iter Iterator) (Func, errors.Error) {
	return Func(iter.Next), nil
}
