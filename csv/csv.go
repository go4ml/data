package csv

import (
	"encoding/csv"
	"io"
	"reflect"
	"go4ml.xyz/data"
	"go4ml.xyz/errstr"
	"go4ml.xyz/fu"
	"go4ml.xyz/iokit"
	"go4ml.xyz/lazy"
	"sync"
)

type Comma rune

const initialCapacity = 101

/*
	// detects compression automatically
    // can be gzip, bzip2, xz/lzma2
	csv.Read(iokit.Compressed(iokit.File("file.csv.xz")),
				csv.Float64("feature_1").As("Feature1"),
				csv.Time("feature_2").Like(time.RFC3339Nano).As("Feature2"))

	// will be downloaded every time
	csv.Read(iokit.Compressed(iokit.Url("s3://$/tests/testfile.csv.xz")))

	// will be downloaded only once
	csv.Read(iokit.Compressed(
				iokit.Url("http://sudachen.xyz/testfile.xz",
					iokit.Cached("external-files/sudachen.xyz/testfile.xz"))))

	// loads file from the Zip archive
	csv.Read(iokit.ZipFile("dataset1.csv",iokit.File("file.zip")))

	csv.Read(iokit.ZipFile("dataset1.csv"
				iokit.Url("http://sudachen.xyz/testfile.zip",
					iokit.Cache("external-files/sudachen.xyz/testfile.zip")))

	var csvContent =
    `s1,f_*,f_1,f_2
  	"the first",100,0,0.1
	"another one",200,3,0.2`

	csv.Read(iokit.StringIO(csvContent),
                csv.TzeInt("f_**").As("Number"), // hide f_* for next rules
				csv.Float64("f_*").As("Feature*"), // => Feature1, Feature2
				csv.String("s*").As("Text*")) // => Text1
*/

type context struct {
	mapper  []mapper
	pool    sync.Pool
	factory data.SimpleRowFactory
}

type value struct {
	F []string
	*context
}

func (v *value) Recycle() {
	v.pool.Put(v)
}

func Read(input interface{}, opts ...interface{}) (t data.Table, err error) {
	err = Source(input, opts...).Drain(t.Sink())
	return
}

func MustRead(input interface{}, opts ...interface{}) (t data.Table) {
	Source(input, opts...).MustDrain(t.Sink())
	return
}

func Source(source interface{}, opts ...interface{}) lazy.Source {
	if e, ok := source.(iokit.Input); ok {
		return read(e, opts...).Map(value2row)
	} else if e, ok := source.(string); ok {
		return read(iokit.File(e), opts...).Map(value2row)
	} else if rd, ok := source.(io.Reader); ok {
		return read(iokit.Reader(rd, nil), opts...).Map(value2row)
	}
	return lazy.ErrorSource(errstr.Errorf("csv reader does not know source type %v", reflect.TypeOf(source).String()))
}

func value2row(x interface{}) interface{} {
	l := x.(*value)
	r := l.factory.New()
	for i, v := range l.F {
		var err error
		m := &l.mapper[i]
		err = m.Convert(v, &r.Data[m.field].Val, m.index, m.width)
		if err != nil {
			return lazy.Fail(err)
		}
	}
	return r
}

func read(source iokit.Input, opts ...interface{}) lazy.Source {
	return func(xs ...interface{}) lazy.Stream {
		worker := 0
		pf := lazy.NoPrefetch

		for _, x := range xs {
			if f, ok := x.(func() (int, int, lazy.Prefetch)); ok {
				worker, _, pf = f()
			} else {
				return lazy.Error(errstr.Errorf("unsupported source option: %v", x))
			}
		}

		return pf(worker, func() lazy.Stream {
			index := 0

			rd, err := source.Open()
			if err != nil {
				return lazy.Error(err)
			}
			cls := io.Closer(rd)
			rdr := csv.NewReader(rd)
			rdr.Comma = fu.RuneOption(Comma(rdr.Comma), opts)
			vals, err := rdr.Read()
			if err != nil {
				_ = cls.Close()
				return lazy.Error(err)
			}
			fm, names, err := mapFields(vals, opts)
			if err != nil {
				_ = cls.Close()
				return lazy.Error(err)
			}

			rdr.FieldsPerRecord = len(vals)
			ctx := &context{mapper: fm}
			ctx.pool.New = func() interface{} { return &value{context: ctx} }
			ctx.factory.Names = names

			return func(next bool) (interface{}, int) {
				if !next {
					_ = cls.Close()
					return lazy.EoS, 0
				}

				row, err := rdr.Read()

				if err != nil {
					if err == io.EOF {
						err = nil
					}
					return lazy.EndOfStream{err}, index
				}

				i := index
				v := ctx.pool.Get().(*value)
				v.F = row
				index++
				return v, i
			}
		})
	}
}
