package data

import "go4ml.xyz/lazy"

type Column struct{ Sequence }

func Col(interface{}) Column {
	return Column{nil}
}

func (c Column) Lazy() lazy.Source {
	return func(...interface{}) lazy.Stream {
		clen := c.Len()
		i := 0
		return func(next bool) (interface{}, int) {
			if next && i < clen {
				j := i
				r := c.At(i).Val
				i++
				return r, j
			}
			return lazy.EoS, i
		}
	}
}
