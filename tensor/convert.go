package tensor

import (
	"reflect"
	"sudachen.xyz/pkg/data"
	"sudachen.xyz/pkg/errstr"
	"sudachen.xyz/pkg/fu"
)

type Xtensor struct{ T reflect.Type }

func (t Xtensor) Type() reflect.Type {
	return fu.TensorType
}

func (t Xtensor) Convert(value string, data *interface{}, index, _ int) (err error) {
	*data, err = DecodeTensor(value)
	return
}

func tensorOf(x interface{}, tp reflect.Type, width int) (data.Tensor, error) {
	if x != nil {
		return x.(data.Tensor), nil
	}
	switch tp {
	case fu.Float32:
		return MakeFloat32Tensor(1, 1, width, nil), nil
	//case fu.Fixed8Type:
	//	return MakeFixed8Tensor(1, 1, width, nil), nil
	//case fu.Int:
	//	return MakeIntTensor(1, 1, width, nil), nil
	case fu.Byte:
		return MakeByteTensor(1, 1, width, nil), nil
	default:
		return data.Tensor{}, errstr.Errorf("unknown tensor value type " + tp.String())
	}
}

func (t Xtensor) ConvertElm(value string, data *interface{}, index, width int) (err error) {
	z, err := tensorOf(*data, t.T, width)
	if err != nil {
		return
	}
	if *data == nil {
		*data = z
	}
	return z.ConvertElem(value, index)
}

func (Xtensor) Format(x interface{}) (string,error) {
	if x == nil {
		return "", nil
	}
	if tz, ok := x.(data.Tensor); ok {
		return tz.String(), nil
	}
	return "", errstr.Errorf("`%v` is not a tensor value", x)
}

