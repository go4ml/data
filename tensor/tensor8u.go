package tensor

import (
	"reflect"
	"strconv"
	"go4ml.xyz/data"
	"go4ml.xyz/fu"
)

type tensor8u struct {
	data.Dim
	values []byte
}

func (t tensor8u) ConvertElem(val string, index int) (err error) {
	i, err := strconv.ParseInt(val, 10, 8)
	if err != nil {
		return
	}
	t.values[index] = byte(i)
	return
}

func (t tensor8u) Index(index int) interface{} { return t.values[index] }

func (t tensor8u) Values() interface{} { return t.values }

func (t tensor8u) Type() reflect.Type { return fu.Byte }

func (t tensor8u) Magic() byte { return 'u' }

func (t tensor8u) HotOne() (j int) {
	for i, v := range t.values {
		if t.values[j] < v {
			j = i
		}
	}
	return
}

func (t tensor8u) CopyTo(r interface{}) {
	out := reflect.ValueOf(r)
	for i, v := range t.values {
		out.Index(i).Set(reflect.ValueOf(v))
	}
}

func (t tensor8u) Floats32(...bool) (r []float32) {
	r = make([]float32, len(t.values))
	for i, v := range t.values {
		r[i] = float32(v) / 256
	}
	return
}

func MakeByteTensor(channels, height, width int, values []byte, docopy ...bool) data.Tensor {
	v := values
	if values != nil {
		if len(docopy) > 0 && docopy[0] {
			v := make([]byte, len(values))
			copy(v, values)
		}
	} else {
		v = make([]byte, channels*height*width)
	}
	x := tensor8u{
		Dim: data.Dim{
			Channels: channels,
			Height:   height,
			Width:    width},
		values: v,
	}
	return data.Tensor{x}
}
