package data

import (
	"reflect"
	"go4ml.xyz/fu"
	"go4ml.xyz/lazy"
	"sync"
)

const defaultMaxVarpartLength = 32*1024/8 - 16

type varpart struct {
	columns                   []reflect.Value
	na                        []fu.Bits
	offset, length, allocated int
}

type varframe struct {
	factory          SimpleRowFactory
	parts            []*varpart
	length           int
	pool             sync.Pool
	maxVarpartLength int
	reserveOnStart   int
}

type varseq struct {
	fr     *varframe
	colndx int
}

func (fr *varframe) last() *varpart {
	if len(fr.parts) == 0 {
		return nil
	}
	return fr.parts[len(fr.parts)-1]
}

func (fr *varframe) append(data []Cell) {
	last := fr.last()
	if last == nil || last.length == fr.maxVarpartLength {
		reserve := fr.reserveOnStart
		if last != nil {
			if len(fr.parts) > 2 {
				reserve = fr.maxVarpartLength
			} else {
				reserve = fr.maxVarpartLength / 2
			}
		}
		width := fr.Width()
		last = &varpart{
			columns:   make([]reflect.Value, width),
			na:        make([]fu.Bits, width),
			offset:    fr.length,
			length:    0,
			allocated: reserve,
		}
		for i := range fr.factory.Names {
			tp := reflect.TypeOf(data[i].Val)
			if !data[i].Na() {
				for tp.Kind() == reflect.Interface {
					tp = tp.Elem()
				}
			}
			last.columns[i] = reflect.MakeSlice(reflect.SliceOf(tp), last.allocated, last.allocated)
		}
		fr.parts = append(fr.parts, last)
	}
	row := last.length
	for i, v := range data {
		if v.Na() {
			last.na[i].Set(row, true)
			if last.length < last.allocated {
				// do nothing, it's already zero
			} else {
				last.columns[i] = reflect.Append(last.columns[i], reflect.Zero(last.columns[i].Type().Elem()))
			}
		} else {
			val := reflect.ValueOf(v.Val)
			for val.Kind() == reflect.Interface {
				val = val.Elem()
			}
			if last.length < last.allocated {
				last.columns[i].Index(last.length).Set(val)
			} else {
				last.columns[i] = reflect.Append(last.columns[i], val)
			}
		}
	}
	last.length++
	fr.length++
}

func (fr *varframe) locate(row int) (*varpart, int) {
	part := row / fr.maxVarpartLength
	if part < len(fr.parts) {
		ndx := row % fr.maxVarpartLength
		if ndx < fr.parts[part].length {
			return fr.parts[part], ndx
		}
	}
	return nil, 0
}

func (fr *varframe) Len() int {
	return fr.length
}

func (fr *varframe) Width() int {
	return fr.factory.Width()
}

func (fr *varframe) Name(col int) string {
	return fr.factory.Name(col)
}

func (fr *varframe) Col(name string) Sequence {
	if colndx, ok := fr.factory.Index(name); ok {
		return &varseq{fr, colndx}
	}
	return nulseq{}
}

func (fr *varframe) At(col int) Sequence {
	if col >= 0 && col < fr.factory.Width() {
		return &varseq{fr, col}
	}
	return nulseq{}
}

func (fr *varframe) Row(index int) *Row {
	row := fr.factory.New()
	part, ndx := fr.locate(index)
	for i, c := range part.columns {
		if !part.na[i].Bit(ndx) {
			row.Data[i].Val = c.Index(ndx).Interface()
		}
	}
	return row
}

func (fr *varframe) Lazy() lazy.Source {
	return lazyFrame(fr)
}

func (sq *varseq) Len() int {
	return sq.fr.length
}

func (sq *varseq) At(index int) Cell {
	part, ndx := sq.fr.locate(index)
	if !part.na[sq.colndx].Bit(ndx) {
		return Cell{part.columns[sq.colndx].Index(ndx).Interface()}
	}
	return Cell{}
}

func (sq *varseq) Na(index int) bool {
	part, ndx := sq.fr.locate(index)
	return part.na[sq.colndx].Bit(ndx)
}

func (sq *varseq) Copy(to interface{}, offset, length int) {
	v := reflect.ValueOf(to)
	for i := 0; i < length; i++ {
		part, ndx := sq.fr.locate(offset + i)
		if !part.na[sq.colndx].Bit(ndx) {
			v.Index(i).Set(part.columns[sq.colndx].Index(ndx))
		}
	}
}

type nulseq struct{}

func (nulseq) Len() int                   { return 0 }
func (nulseq) At(index int) Cell          { return Cell{} }
func (nulseq) Na(index int) bool          { return true }
func (nulseq) Copy(interface{}, int, int) {}
