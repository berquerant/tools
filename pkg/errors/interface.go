/*
Package errors provides extended errors
*/
package errors

import "fmt"

type (
	// Error is extended error
	Error interface {
		error
		SetCode(c Code) Error
		SetError(err error) Error
	}

	xError struct {
		C Code
		E error
	}
)

//go:generate stringer -type=Code -output generated.code_string.go
type Code int

const (
	// Unknown is unknown termination
	Unknown Code = iota
	// Normal is normal termiantion
	Normal
	// System is from system
	System
	// Parse is parse error
	Parse
	// Translate is translate error
	Translate
	// IO is io error
	IO
	// Validate is validate error
	Validate
	// Iterator is error by iterator
	Iterator
	// Conversion is type conversion error
	Conversion
	// Fold is fold error
	Fold
	// Map is map error
	Map
	// Filter is filter error
	Filter
	// Consume is consume error
	Consume
	// Sort is sort error
	Sort
	// Lift is lift error
	Lift
)

func NewError() Error {
	return &xError{}
}

func (s *xError) SetError(err error) Error {
	s.E = err
	return s
}

func (s *xError) SetCode(c Code) Error {
	s.C = c
	return s
}

func (s *xError) Error() string {
	return fmt.Sprintf("%v %v", s.C, s.E)
}
