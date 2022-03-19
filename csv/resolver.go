package csv

import (
	"fmt"
	"strconv"
	"go4ml.xyz/data"
	"go4ml.xyz/data/tensor"
	"go4ml.xyz/fu"
	"time"
)

type Resolver func() mapper

func (r Resolver) As(n string) Resolver {
	return func() mapper {
		m := r()
		m.TableCol = n
		return m
	}
}

func Column(v string) Resolver {
	return func() mapper {
		return mapAs(v, v, nil, nil, nil)
	}
}

func (r Resolver) Group(v string) Resolver {
	return func() mapper {
		g := r()
		z := tensor.Xtensor{g.valueType}
		x := g
		x.TableCol = v
		x.group = true
		x.valueType = z.Type()
		x.convert = func(value string, data *interface{}, index, width int) error {
			return z.ConvertElm(value, data, index, width)
		}
		return x
	}
}

func Tensor32f(v string) Resolver {
	return func() mapper {
		x := tensor.Xtensor{fu.Float32}
		return mapAs(v, v, x.Type(), x.Convert, x.Format)
	}
}

func Tensor64f(v string) Resolver {
	return func() mapper {
		x := tensor.Xtensor{fu.Float64}
		return mapAs(v, v, x.Type(), x.Convert, x.Format)
	}
}

func Tensor32i(v string) Resolver {
	return func() mapper {
		x := tensor.Xtensor{fu.Int32}
		return mapAs(v, v, x.Type(), x.Convert, x.Format)
	}
}

func Tensor8u(v string) Resolver {
	return func() mapper {
		x := tensor.Xtensor{fu.Byte}
		return mapAs(v, v, x.Type(), x.Convert, x.Format)
	}
}

func Tensor8f(v string) Resolver {
	return func() mapper {
		x := tensor.Xtensor{fu.Fixed8Type}
		return mapAs(v, v, x.Type(), x.Convert, x.Format)
	}
}

func Meta(x data.Meta, v string) Resolver {
	return func() mapper {
		return mapAs(v, v, x.Type(), x.Convert, x.Format)
	}
}

func String(v string) Resolver {
	return func() mapper {
		return mapAs(v, v, fu.String, nil, nil)
	}
}

func Int(v string) Resolver {
	return func() mapper {
		return mapAs(v, v, fu.Int, converti, nil)
	}
}

func converti(s string, data *interface{}, _, _ int) (err error) {
	if s == "" {
		*data = nil
		return
	}
	*data, err = strconv.ParseInt(s, 10, 64)
	return
}

func Fixed8(v string) Resolver {
	return func() mapper {
		return mapAs(v, v, fu.Fixed8Type, convert8f, nil)
	}
}

func convert8f(s string, data *interface{}, _, _ int) (err error) {
	if s == "" {
		*data = nil
		return
	}
	*data, err = fu.Fast8f(s)
	return
}

func Float32(v string) Resolver {
	return func() mapper {
		return mapAs(v, v, fu.Float32, convert32f, nil)
	}
}

func convert32f(s string, data *interface{}, _, _ int) (err error) {
	if s == "" {
		*data = float32(0)
		return
	}
	*data, err = fu.Fast32f(s)
	return
}

func Float64(v string) Resolver {
	return func() mapper {
		return mapAs(v, v, fu.Float64, convert64f, nil)
	}
}

func convert64f(s string, data *interface{}, _, _ int) (err error) {
	if s == "" {
		*data = nil
		return
	}
	*data, err = strconv.ParseFloat(s, 32)
	return
}

func Time(v string, layout ...string) Resolver {
	l := time.RFC3339
	if len(layout) > 0 {
		l = layout[0]
	}
	return func() mapper {
		return mapAs(v, v, fu.Ts,
			func(s string, data *interface{}, i, _ int) (error) {
				return convertts(s, l, data)
			}, nil)
	}
}

func convertts(s string, layout string, value *interface{}) (err error) {
	if s == "" {
		*value = nil
		return
	}
	//v, err := ???
	return
}

func (r Resolver) Round(n ...int) Resolver {
	return func() mapper {
		m := r()
		xf := m.format
		m.format = func(v interface{}) (string, error) {
			if v != nil {
				switch fv := v.(type) {
				case float64:
					v = fu.Round64(fv, n[0])
				case float32:
					v = fu.Round32(fv, n[0])
				}
			}
			return format(v, xf)
		}
		return m
	}
}

func format(v interface{}, xf formatter) (string, error) {
	if xf != nil {
		return xf(v)
	}
	return fmt.Sprint(v), nil
}
