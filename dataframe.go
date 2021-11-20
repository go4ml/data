package data

import "sudachen.xyz/pkg/lazy"

type Frame interface {
	Len() int
	Width() int
	Name(int) string
	At(column int) Sequence
	Col(column string) Sequence
	Row(index int) *Row
	Lazy() lazy.Source
}

type Sequence interface {
	Len() int
	At(index int) Cell
	Na(index int) bool
	Copy(to interface{}, offset, length int)
}
