package data

import "go4ml.xyz/fu"

type tableSlice struct {
	Frame
	offset, length int
}

type seqSlice struct {
	Sequence
	offset, length int
}

func (t Table) Slice(offset, length int) Table {
	if sl, ok := t.Frame.(*tableSlice); ok {
		newOffset := fu.Mini(sl.offset+offset, sl.offset+sl.length)
		newLength := fu.Maxi(0, sl.length-offset)
		return Table{&tableSlice{sl.Frame, newOffset, newLength}}
	}
	return Table{&tableSlice{
		t.Frame,
		offset,
		fu.Maxi(0, fu.Mini(t.Len()-offset, length))}}
}

func (sl tableSlice) Len() int {
	return sl.length
}

func (sl tableSlice) Column(n string) Sequence {
	return &seqSlice{
		sl.Frame.Col(n),
		sl.offset,
		sl.length}
}

func (sl tableSlice) Row(index int) *Row {
	return sl.Frame.Row(sl.offset + index)
}

func (sl seqSlice) Len() int {
	return sl.length
}

func (sl seqSlice) At(i int) Cell {
	return sl.Sequence.At(sl.offset + i)
}

func (sl seqSlice) Na(i int) bool {
	return sl.Sequence.Na(sl.offset + i)
}

func (sl seqSlice) Copy(to interface{}, offset, length int) {
	sl.Sequence.Copy(to, sl.offset+offset, fu.Maxi(0, fu.Mini(length, sl.length-offset)))
}
