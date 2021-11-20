package data

import (
	"sort"
	"sync"
)

type RowFactory interface {
	New() *Row
	Recycle(*Row)
	Index(n string) (int, bool)
	Name(int) string
	With(with []string) RowFactory
	Except(except []string) RowFactory
	Width() int
}

type SimpleRowFactory struct {
	Pool  sync.Pool
	Names []string
}

func (f *SimpleRowFactory) Init(names []string) {
	f.Names = make([]string, len(names))
	copy(f.Names, names)
}

func (f *SimpleRowFactory) InitFrom(from RowFactory) {
	f.Names = make([]string, from.Width())
	for i := 0; i < len(f.Names); i++ {
		f.Names[i] = from.Name(i)
	}
}

func NewRowFactory(names []string) *SimpleRowFactory {
	f := &SimpleRowFactory{}
	f.Init(names)
	return f
}

func (f *SimpleRowFactory) New() *Row {
	if x := f.Pool.Get(); x != nil {
		return x.(*Row)
	}
	return &Row{f, make([]Cell, len(f.Names))}
}

func (f *SimpleRowFactory) With(with []string) RowFactory {
	ns := make([]string, len(f.Names), len(with))
	copy(ns, f.Names)
	for _, n := range with {
		j := sort.SearchStrings(ns, n)
		if ns[j] != n {
			k := len(ns)
			ns = append(ns, n)
			copy(ns[j+1:], ns[j:k])
			ns[j] = n
		}
	}
	return &SimpleRowFactory{Pool: sync.Pool{}, Names: ns}
}

func (f *SimpleRowFactory) Except(except []string) RowFactory {
	ns := make([]string, 0, len(f.Names))
loop:
	for _, n := range f.Names {
		for _, e := range except {
			if e == n {
				continue loop
			}
		}
		ns = append(ns, n)
	}
	return &SimpleRowFactory{Pool: sync.Pool{}, Names: ns}
}

func (f *SimpleRowFactory) Recycle(r *Row) {
	for i := range r.Data {
		r.Data[i] = Cell{nil}
	}
	f.Pool.Put(r)
}

func (f *SimpleRowFactory) Index(n string) (int, bool) {
	for i, s := range f.Names {
		if s == n {
			return i, true
		}
	}
	return -1, false
}

func (f *SimpleRowFactory) Name(n int) string {
	if n >= 0 && n < len(f.Names) {
		return f.Names[n]
	}
	return ""
}

func (f *SimpleRowFactory) Width() int {
	return len(f.Names)
}
