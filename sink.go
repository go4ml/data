package data

import (
	"sudachen.xyz/pkg/fu"
	"sudachen.xyz/pkg/lazy"
)

func sink1(t *Table, res int) []lazy.Worker {
	var fr *varframe
	return []lazy.Worker{func(_ int, v interface{}, err error) (_ error) {
		if v != nil {
			switch r := v.(type) {
			case *Row:
				if fr == nil {
					fr = &varframe{
						maxVarpartLength: fu.Maxi(defaultMaxVarpartLength, res),
						reserveOnStart:   res,
					}
					fr.factory.InitFrom(r.Factory)
				}
				fr.append(r.Data)
				r.Recycle()
			}
		} else if err == nil {
			*t = Table{fr}
		}
		return nil
	}}
}

func sink2(t *Table, res int, concurrency int) []lazy.Worker{
	wf := make([]lazy.Worker, concurrency)
	vf := make([]*varframe, concurrency)
	ndx := make([][]int, concurrency)
	res = res/concurrency + res/10
	for k := range wf {
		wf[k] = func(w int, fr *varframe)lazy.Worker{
			return func(i int, v interface{}, e error)error{
				if v != nil {
					switch r := v.(type) {
					case *Row:
						if fr == nil {
							fr = &varframe{
								maxVarpartLength: fu.Maxi(defaultMaxVarpartLength,res),
								reserveOnStart:   res,
							}
							fr.factory.InitFrom(r.Factory)
							vf[w] = fr
						}
						fr.append(r.Data)
						ndx[w] = append(ndx[w],i)
						r.Recycle()
					}
				} else if e == nil {
					*t = Table{ ccrComplete(vf, ndx) }
				}
				return nil
			}
		}(k, nil)
	}
	return wf
}

func (t *Table) Sink(reserve ...int) lazy.WorkerFactory {
	res := fu.Fnzi(reserve...)
	return func(concurrency int) []lazy.Worker {
		if concurrency < 2 {
			return sink1(t, res)
		}
		return sink2(t, res, concurrency)
	}
}
