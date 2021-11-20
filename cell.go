package data

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sudachen.xyz/pkg/errstr"
)

type Cell struct {
	Val interface{}
}

func (c Cell) Type() reflect.Type {
	return reflect.TypeOf(c.Val)
}

func (c Cell) Na() bool {
	return c.Val == nil
}

func (c Cell) Text() string {
	switch v := c.Val.(type) {
	case string:
		return v
	case Tensor:
		var s []string
		dim := v.Dimension()
		for i := 0; i < 4; i++ {
			if i == 3 || i >= dim.Volume() {
				s = append(s, ">")
				break
			} else if i < dim.Volume() {
				s = append(s, fmt.Sprint(v.Index(i)))
			}
		}
		ch, h, w := dim.CHW()
		return fmt.Sprintf("(%dx%dx%d){%v}", ch, h, w, strings.Join(s, ","))
	default:
		return fmt.Sprint(v)
	}
}

func (c Cell) String() string { return c.Text() }

func (c Cell) Tensor() Tensor {
	switch v := c.Val.(type) {
	case Tensor:
		return v
	default:
		panic(errstr.Errorf("can't convert %v to Tensor", reflect.TypeOf(c.Val)))
	}
}

func (c Cell) Int() int {
	switch v := c.Val.(type) {
	case int:
		return v
	default:
		return int(c.Int64())
	}
}

func (c Cell) Byte() byte {
	switch v := c.Val.(type) {
	case byte:
		return v
	default:
		return byte(c.Uint64())
	}
}

func (c Cell) float32() float32 {
	switch v := c.Val.(type) {
	case float32:
		return v
	default:
		return float32(c.Float64())
	}
}

func (c Cell) Int64() int64 {
	switch v := c.Val.(type) {
	case uint8:
		return int64(v)
	case int8:
		return int64(v)
	case int:
		return int64(v)
	case uint:
		return int64(v)
	case int16:
		return int64(v)
	case uint16:
		return int64(v)
	case int32:
		return int64(v)
	case uint32:
		return int64(v)
	case int64:
		return v
	case uint64:
		return int64(v)
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	case string:
		x, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			panic(err)
		}
		return x
	default:
		if c.Na() {
			return 0
		}
		panic(errstr.Errorf("can't convert %v to int", reflect.TypeOf(c.Val)))
	}
}

func (c Cell) Uint64() uint64 {
	switch v := c.Val.(type) {
	case uint8:
		return uint64(v)
	case int8:
		return uint64(v)
	case int:
		return uint64(v)
	case uint:
		return uint64(v)
	case int16:
		return uint64(v)
	case uint16:
		return uint64(v)
	case int32:
		return uint64(v)
	case uint32:
		return uint64(v)
	case int64:
		return uint64(v)
	case uint64:
		return v
	case float32:
		return uint64(v)
	case float64:
		return uint64(v)
	case string:
		x, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			panic(err)
		}
		return x
	default:
		if c.Na() {
			return 0
		}
		panic(errstr.Errorf("can't convert %v to int", reflect.TypeOf(c.Val)))
	}
}

func (c Cell) Float64() float64 {
	switch v := c.Val.(type) {
	case uint8:
		return float64(v)
	case int8:
		return float64(v)
	case int:
		return float64(v)
	case uint:
		return float64(v)
	case int16:
		return float64(v)
	case uint16:
		return float64(v)
	case int32:
		return float64(v)
	case uint32:
		return float64(v)
	case int64:
		return float64(v)
	case uint64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	case string:
		x, err := strconv.ParseFloat(v, 64)
		if err != nil {
			panic(err)
		}
		return x
	default:
		if c.Na() {
			return math.NaN()
		}
		panic(errstr.Errorf("can't convert %v to int", reflect.TypeOf(c.Val)))
	}
}
