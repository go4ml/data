package data

import (
	"fmt"
	"reflect"
	"strings"
)

type Row struct {
	Factory RowFactory
	Data    []Cell
}

func (r *Row) Recycle() {
	r.Factory.Recycle(r)
}

func (r *Row) Width() int {
	return r.Factory.Width()
}

func (r *Row) Index(name string) (int, bool) {
	return r.Factory.Index(name)
}

func (r *Row) At(i int) Cell {
	if i >= 0 && i < len(r.Data) {
		return r.Data[i]
	}
	return Cell{}
}

func (r *Row) Col(n string) Cell {
	if i, ok := r.Index(n); ok {
		return r.Data[i]
	}
	return Cell{}
}

func (r Row) String() string {
	var ss []string
	for i := 0; i < len(r.Data); i++ {
		n := r.Factory.Name(i)
		ss = append(ss, fmt.Sprintf("%s: %v(%v)", n, reflect.TypeOf(r.Data[i].Val), r.Data[i].Val))
	}
	return "Row{" + strings.Join(ss, ", ") + "}"
}
