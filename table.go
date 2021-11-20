package data

import (
	"sudachen.xyz/pkg/lazy"
)

type Table struct{ Frame }

func (t Table) At(col int) Column {
	return Column{t.Frame.At(col)}
}

func (t Table) Col(name string) Column {
	return Column{t.Frame.Col(name)}
}

func (t Table) Lazy() lazy.Source {
	return t.Frame.Lazy()
}

func (t Table) Sort(less func(Table,int,int)bool) Table {
	return Table{}
}

func (t Table) SortBy(col ...string) Table {
	return Table{}
}

func lazyFrame(fr Frame) lazy.Source {
	return func(...interface{}) lazy.Stream {
		clen := fr.Len()
		i := 0
		return func(next bool) (interface{},int) {
			if i < clen && next {
				j := i
				r := fr.Row(i)
				i++
				return r, j
			}
			return lazy.EoS, 0
		}
	}
}
