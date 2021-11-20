package data

import (
	"reflect"
	"sudachen.xyz/pkg/errstr"
	"sudachen.xyz/pkg/lazy"
)

type StructWrapper struct {
	SimpleRowFactory
	tp     reflect.Type
	ndxmap []int
}

func NewWrapper(st interface{}) (w *StructWrapper, err error) {
	tp := reflect.TypeOf(st)
	if tp.Kind() == reflect.Interface {
		tp = tp.Elem()
	}
	w = &StructWrapper{tp: tp}
	flen := tp.NumField()
	w.ndxmap = make([]int, flen)
	w.SimpleRowFactory.Names = make([]string, flen)
	for i := 0; i < flen; i++ {
		w.ndxmap[i] = i
		w.SimpleRowFactory.Names[i] = tp.Field(i).Name
	}
	return
}

func (w *StructWrapper) Wrap(st interface{}) (row *Row, err error) {
	row = w.New()
	if err = w.Fill(row, st); err != nil {
		row.Recycle()
		row = nil
	}
	return
}

func (w *StructWrapper) WrapOrFail(st interface{}) interface{} {
	row := w.New()
	if err := w.Fill(row, st); err != nil {
		row.Recycle()
		return lazy.Fail(err)
	}
	return row
}

func (w *StructWrapper) Fill(row *Row, st interface{}) (_ error) {
	v := reflect.ValueOf(st)
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if v.Type() != w.tp {
		return errstr.Errorf("can't fill row of %v from %v", w.tp, v.Type())
	}
	for from, to := range w.ndxmap {
		row.Data[to].Val = v.Field(from).Interface()
	}
	return
}

func StructToRow(st interface{}) interface{} {
	w, err := NewWrapper(st)
	return func(x interface{}) interface{} {
		var v interface{}
		if err == nil {
			v, err = w.Wrap(x)
		}
		if err != nil {
			return lazy.Fail(err)
		}
		return v
	}
}

func ValueToRow(col string) interface{} {
	return func(st interface{}) interface{} {
		factory := SimpleRowFactory{Names: []string{col}}
		return func(x interface{}) interface{} {
			row := factory.New()
			row.Data[0].Val = x
			return row
		}
	}
}
