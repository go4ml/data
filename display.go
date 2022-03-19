package data

import (
	"fmt"
	"go4ml.xyz/fu"
	"strings"
)

func (t Table) Head(n int) string {
	if n < t.Len() {
		return t.Display(0, n)
	}
	return t.Display(0, t.Len())
}

func (t Table) Tail(n int) string {
	ln := t.Len()
	if n < ln {
		return t.Display(ln-n, ln)
	}
	return t.Display(0, ln)
}

func (t Table) String() string {
	return t.Display(0, t.Frame.Len())
}

func (t Table) Display(from, to int) string {
	ln := t.Len()
	n := fu.Mini(to-from, ln-from)
	if n < 0 {
		n = 0
	}
	w := make([]int, t.Width()+1)
	s := make([][]interface{}, n+1)
	s[0] = append(make([]interface{}, len(w)))
	s[0][0] = ""
	for i := 0; i < t.Width(); i++ {
		name := t.Name(i)
		s[0][i+1] = name
		w[i+1] = len(name)
	}
	for k := 0; k < n; k++ {
		u := make([]interface{}, len(w))
		ln := fmt.Sprint(k + from)
		if w[0] < len(ln) {
			w[0] = len(ln)
		}
		u[0] = ln
		for j := range w[1:] {
			ws := t.At(j).At(k + from).String()
			if len(ws) > w[j+1] {
				w[j+1] = len(ws)
			}
			u[j+1] = ws
		}
		s[k+1] = u
	}
	f0 := ""
	f1 := ""
	f2 := ""
	for i, v := range w {
		if i != 0 {
			f0 += " . "
			f1 += " | "
			f2 += "-|-"
		}
		q := fmt.Sprintf("%%-%ds", v)
		f0 += q
		f1 += q
		f2 += strings.Repeat("-", v)
	}
	r := ""
	r += fmt.Sprintf(f0, s[0]...) + "\n"
	r += f2 + "\n"
	for _, u := range s[1:] {
		r += fmt.Sprintf(f1, u...) + "\n"
	}
	return r[:len(r)-1]
}
