package csv

import (
	"encoding/csv"
	"go4ml.xyz/data"
	"go4ml.xyz/errstr"
	"go4ml.xyz/fu"
	"go4ml.xyz/iokit"
	"go4ml.xyz/lazy"
)

/*
	csv.Write(t,iokit.File("file.csv.xz"),
				csv.Column("feature_1").Round(2).As("Feature1"))

	csv.Write(t,iokit.LzmaFile("file.csv.xz"),
				csv.Column("feature*").As("Feature*"))

	bf := bytes.Buffer{}
	csv.Write(t,iokit.GzipWriter(&bf),
				csv.Comma('|'),
				csv.Column("feature*").Round(3).As("Feature*"))

	csv.Write(t,iokit.LzmaUrl("gc://$/testfile.csv.xz"),
				csv.Comma('|'),
				csv.Column("feature_1").As("Feature1"))
*/

func Write(t data.Table, dest iokit.Output, opts ...interface{}) (err error) {
	return t.Lazy().Drain(Sink(dest, opts...))
}

func Sink(dest iokit.Output, opts ...interface{}) lazy.WorkerFactory {
	var err error
	f := iokit.Whole(nil)
	if f, err = dest.Create(); err != nil {
		return lazy.ErrorSink(err)
	}
	cwr := csv.NewWriter(f)
	hasHeader := false
	fm := []mapper{}
	names := []string{}

	// synchronous write
	return lazy.Sink(func(v interface{}, err error) error {
		if v == nil {
			cwr.Flush()
			if err == nil {
				err = f.Commit()
			}
			f.End()
			return err
		}
		switch x := v.(type) {
		case struct{}:
			// skip
			return nil
		case *data.Row:
			if !hasHeader {
				names = make([]string, x.Factory.Width())
				for i := range names {
					names[i] = x.Factory.Name(i)
				}
				if fm, names, err = mapFields(names, opts); err != nil {
					return err
				}
				if err = cwr.Write(names); err != nil {
					return err
				}
				hasHeader = true
			}
			r := make([]string, len(names))
			for i, _ := range r {
				if r[i], err = fm[i].Format(x.Data[i].Val); err != nil {
					return err
				}
			}
			return cwr.Write(r)
		default:
			return errstr.Errorf("unsupported value type %v", fu.TypeOf(v))
		}
	})
}
