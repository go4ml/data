package data

import "go4ml.xyz/lazy"

type xrow struct{ frame, row uint32 }
type ccrframe struct {
	vf     []*varframe
	index  []xrow
	length int
}

func ccrComplete(vf []*varframe, ndx [][]int) Frame {
	const Barrier = int(^uint(0) >> 1)

	length := 0
	for _, x := range ndx {
		length += len(x)
	}

	index := make([]xrow, length)

	type yrow struct{ f, n, ln, i int }
	h := make([]yrow, len(ndx))

	for f, x := range ndx {
		if ln := len(x); ln > 0 {
			h[f] = yrow{f, x[0], ln, 0}
		} else {
			h[f].n = Barrier
		}
	}
	w := len(h)
	i, j, x := 0, w-1, 0
	for {
		jj := (j + 1) % w
		if h[jj].n == h[j].n+1 {
			j = jj
		} else {
			j, x = Barrier, Barrier
			for k, y := range h {
				if y.n < x {
					x = y.n
					j = k
				}
			}
			if j == Barrier {
				break
			}
		}
		index[i] = xrow{uint32(h[j].f), uint32(h[j].i)}
		i++
		if h[j].ln > h[j].i+1 {
			h[j].n = ndx[h[j].f][h[j].i]
			h[j].i++
		} else {
			h[j].n = Barrier
		}
	}
	return &ccrframe{vf: vf, length: length, index: index}
}

func (fr *ccrframe) Len() int {
	return fr.length
}

func (fr *ccrframe) Width() int {
	return fr.vf[0].factory.Width()
}

func (fr *ccrframe) Name(i int) string {
	return fr.vf[0].factory.Name(i)
}

func (fr *ccrframe) At(i int) Sequence {
	if i >= 0 && fr.length > 0 {
		vf := fr.vf[fr.index[0].frame]
		if i < vf.factory.Width() {
			return &ccrseq{fr, i}
		}
	}
	return &nulseq{}
}

func (fr *ccrframe) Col(n string) Sequence {
	if fr.length > 0 {
		vf := fr.vf[fr.index[0].frame]
		if i, ok := vf.factory.Index(n); ok {
			return &ccrseq{fr, i}
		}
	}
	return &nulseq{}
}

func (fr *ccrframe) Row(i int) *Row {
	return fr.vf[fr.index[i].frame].Row(int(fr.index[i].row))
}

func (fr *ccrframe) Lazy() lazy.Source {
	return lazyFrame(fr)
}

type ccrseq struct {
	fr  *ccrframe
	col int
}

func (sq *ccrseq) Len() int {
	return sq.fr.length
}

func (sq *ccrseq) At(i int) Cell {
	fr := sq.fr
	part, ndx := fr.vf[fr.index[i].frame].locate(int(fr.index[i].row))
	if !part.na[sq.col].Bit(ndx) {
		return Cell{part.columns[sq.col].Index(ndx).Interface()}
	}
	return Cell{}
}

func (sq *ccrseq) Na(i int) bool {
	fr := sq.fr
	part, ndx := fr.vf[fr.index[i].frame].locate(int(fr.index[i].row))
	return part.na[sq.col].Bit(ndx)
}

func (sq *ccrseq) Copy(to interface{}, offset, length int) {

}
