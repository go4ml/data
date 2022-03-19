package tensor

import (
	"reflect"
	"go4ml.xyz/data"
	"go4ml.xyz/fu"
)

type tensor32f struct {
	data.Dim
	values []float32
}

func (t tensor32f) ConvertElem(val string, index int) (err error) {
	t.values[index], err = fu.Fast32f(val)
	return
}

func (t tensor32f) Index(index int) interface{} { return t.values[index] }
func (t tensor32f) Values() interface{}         { return t.values }
func (t tensor32f) Type() reflect.Type          { return fu.Float32 }
func (t tensor32f) Magic() byte                 { return 'f' }
func (t tensor32f) HotOne() (j int) {
	for i, v := range t.values {
		if t.values[j] < v {
			j = i
		}
	}
	return
}

func (t tensor32f) CopyTo(r interface{}) {
	out := reflect.ValueOf(r)
	for i, v := range t.values {
		out.Index(i).Set(reflect.ValueOf(v))
	}
}

func (t tensor32f) Floats32(c ...bool) []float32 {
	if fu.Fnzb(c...) {
		r := make([]float32, len(t.values))
		copy(r, t.values)
		return r
	}
	return t.values
}

func MakeFloat32Tensor(channels, height, width int, values []float32, docopy ...bool) data.Tensor {
	v := values
	if values != nil {
		if len(docopy) > 0 && docopy[0] {
			v := make([]float32, len(values))
			copy(v, values)
		}
	} else {
		v = make([]float32, channels*height*width)
	}
	if width <= 0 {
		width = len(values) / (channels * height)
	}
	x := tensor32f{
		Dim: data.Dim{
			Channels: channels,
			Height:   height,
			Width:    width,
		},
		values: v,
	}
	return data.Tensor{x}
}
