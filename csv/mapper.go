package csv

import (
	"reflect"
	"sudachen.xyz/pkg/errstr"
	"sudachen.xyz/pkg/fu"
)

type formatter func(value interface{}) (string,error)
type converter func(value string, data *interface{}, index, width int) error
type mapper struct {
	CsvCol, TableCol string
	valueType        reflect.Type
	convert          converter
	format           formatter
	group            bool
	field, index     int
	width            int
	name             string
}

func mapAs(ccol, tcol string, t reflect.Type, conv converter, form formatter) mapper {
	return mapper{ccol, tcol, t, conv, form, false, 0, 0, 0, ""}
}

func (m mapper) Group() bool {
	return m.group
}

func (m mapper) Type() reflect.Type {
	if m.valueType != reflect.Type(nil) {
		return m.valueType
	}
	return fu.String
}

func (m mapper) Convert(value string, data *interface{}, index, width int) (err error) {
	if m.convert != nil {
		return m.convert(value, data, index, width)
	}
	return nil
}

func (m mapper) Format(v interface{}) (string, error) {
	return format(v, m.format)
}

func mapFields(header []string, opts []interface{}) (fm []mapper, names []string, err error) {
	fm = make([]mapper, len(header))
	names = make([]string, 0, len(header))
	mask := fu.Bits{}
	for _, o := range opts {
		if x, ok := o.(Resolver); ok {
			v := x()
			exists := false
			if v.group {
				like := fu.Pattern(v.CsvCol)
				for i, n := range header {
					if !mask.Bit(i) && like(n) {
						v.name = v.TableCol
						fm[i] = v
						mask.Set(i, true)
						exists = true
						v.index++
					}
				}
			} else {
				starsub := fu.Starsub(v.CsvCol, v.TableCol)
				for i, n := range header {
					if !mask.Bit(i) {
						if c, ok := starsub(n); ok {
							v.name = c
							fm[i] = v
							exists = true
						}
					}
				}
			}
			if !exists {
				return nil, nil, errstr.Errorf("field %v does not exist in CSV file", v.CsvCol)
			}
		}
	}
	width := make([]int, len(header))
	for i := range fm {
		if fm[i].name == "" {
			fm[i].name = header[i]
		}
		j := fu.IndexOf(fm[i].name, names)
		if j < 0 {
			j = len(names)
			names = append(names, fm[i].name)
		}
		fm[i].field = j
		width[j]++
	}
	for i := range fm {
		if fm[i].group {
			fm[i].width = width[fm[i].field]
		}
	}
	return
}
