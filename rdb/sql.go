package rdb

import (
	"database/sql"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sudachen.xyz/pkg/data"
	"sudachen.xyz/pkg/errstr"
	"sudachen.xyz/pkg/fu"
	"sudachen.xyz/pkg/iokit"
	"sudachen.xyz/pkg/lazy"
	//	_ "github.com/go-sql-driver/mysql"
	//	_ "github.com/lib/pq"
	//	_ "github.com/mattn/go-sqlite3"
)

func Read(dbc interface{}, opts ...interface{}) (t data.Table, err error) {
	err = Source(dbc, opts...).Drain(t.Sink())
	return
}

func Write(t data.Table, dbc interface{}, opts ...interface{}) error {
	return t.Lazy().Drain(Sink(dbc, opts...))
}

type dontclose bool

func connectDB(dbc interface{}, opts []interface{}) (db *sql.DB, o []interface{}, err error) {
	o = opts
	if url, ok := dbc.(string); ok {
		drv, conn := splitDriver(url)
		o = append(o, Driver(drv))
		db, err = sql.Open(drv, conn)
	} else if db, ok = dbc.(*sql.DB); !ok {
		err = errstr.Errorf("unknown database %v", dbc)
	} else {
		o = append(o, dontclose(true))
	}
	return
}

func Source(dbc interface{}, opts ...interface{}) lazy.Source {
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

			db, opts, err := connectDB(dbc, opts)
			cls := io.Closer(iokit.CloserChain{})
			if !fu.BoolOption(dontclose(false), opts) {
				cls = db
			}
			if err != nil {
				lazy.Error(errstr.Wrapf(0, err, "database connection error: %s", err.Error()))
			}
			drv := fu.StrOption(Driver(""), opts)
			schema := fu.StrOption(Schema(""), opts)
			if schema != "" {
				switch drv {
				case "mysql":
					_, err = db.Exec("use " + schema)
				case "postgres":
					_, err = db.Exec("set search_path to " + schema)
				}
			}
			if err != nil {
				cls.Close()
				return lazy.Error(errstr.Wrapf(0, err, "query error: %s", err.Error()))
			}
			query := fu.StrOption(Query(""), opts)
			if query == "" {
				table := fu.StrOption(Table(""), opts)
				if table != "" {
					query = "select * from " + table
				} else {
					panic("there is no query or table")
				}
			}
			rows, err := db.Query(query)
			if err != nil {
				cls.Close()
				return lazy.Error(errstr.Wrapf(0, err, "query error: %s", err.Error()))
			}
			cls = iokit.CloserChain{rows, cls}
			tps, err := rows.ColumnTypes()
			if err != nil {
				cls.Close()
				return lazy.Error(errstr.Wrapf(0, err, "get types error: %s", err.Error()))
			}
			ns, err := rows.Columns()
			if err != nil {
				cls.Close()
				return lazy.Error(errstr.Wrapf(0, err, "get names error: %s", err.Error()))
			}
			x := make([]interface{}, len(ns))
			describe, err := Describe(ns, opts)
			if err != nil {
				cls.Close()
				return lazy.Error(err)
			}
			names := make([]string, len(ns))
			for i, n := range ns {
				var s SqlScan
				colType, colName, _ := describe(n)
				if colType != "" {
					s = scanner(colType)
				} else {
					s = scanner(tps[i].DatabaseTypeName())
				}
				x[i] = s
				names[i] = colName
			}

			factory := data.SimpleRowFactory{Names: names}

			return func(next bool) (interface{}, int) {
				if !next {
					_ = cls.Close()
					return lazy.EoS, index
				}
				if rows.Next() {
					e := rows.Scan(x...)
					if e != nil {
						return lazy.EndOfStream{e}, index
					}
					r := factory.New()
					for i := range x {
						y := x[i].(SqlScan)
						if v, ok := y.Value(); ok {
							r.Data[i].Val = v
						}
					}
					j := index
					index++
					return  r, j
				}
				return lazy.EndOfStream{rows.Err()}, index
			}
		})
	}
}

func splitDriver(url string) (string, string) {
	q := strings.SplitN(url, ":", 2)
	return q[0], q[1]
}

func scanner(q string) SqlScan {
	switch q {
	case "VARCHAR", "TEXT", "CHAR", "STRING":
		return &SqlString{}
	case "INT8", "SMALLINT", "INT2":
		return &SqlSmall{}
	case "INTEGER", "INT", "INT4":
		return &SqlInteger{}
	case "BIGINT":
		return &SqlBigint{}
	case "BOOLEAN":
		return &SqlBool{}
	case "DECIMAL", "NUMERIC", "REAL", "DOUBLE", "FLOAT8":
		return &SqlDouble{}
	case "FLOAT", "FLOAT4":
		return &SqlFloat{}
	case "DATE", "DATETIME", "TIMESTAMP":
		return &SqlTimestamp{}
	default:
		if strings.Index(q, "VARCHAR(") == 0 ||
			strings.Index(q, "CHAR(") == 0 {
			return &SqlString{}
		}
		if strings.Index(q, "DECIMAL(") == 0 ||
			strings.Index(q, "NUMERIC(") == 0 {
			return &SqlDouble{}
		}
	}
	panic("unknown column type " + q)
}

func batchInsertStmt(tx *sql.Tx, names []string, pk []bool, lines int, table string, opts []interface{}) (stmt *sql.Stmt, err error) {
	drv := fu.StrOption(Driver(""), opts)
	ifExists := fu.Option(ErrorIfExists, opts).Interface().(IfExists_)
	L := len(names)
	q1 := " values "
	for j := 0; j < lines; j++ {
		q1 += "("
		if drv == "postgres" {
			for k := range names {
				q1 += fmt.Sprintf("$%d,", j*L+k+1)
			}
		} else {
			q1 += strings.Repeat("?,", L)
		}
		q1 = q1[:len(q1)-1] + "),"
	}
	q := "insert into " + table + "(" + strings.Join(names, ",") + ")" + q1[:len(q1)-1]

	if ifExists == InsertUpdateIfExists {
		if len(pk) > 0 {
			q += " on duplicate key update "
			for i, n := range names {
				if !pk[i] {
					q += " " + n + " = values(" + n + "),"
				}
			}
			q = q[:len(q)-1]
		}
	}
	stmt, err = tx.Prepare(q)
	return
}

func Sink(dbc interface{}, opts ...interface{}) lazy.WorkerFactory {
	db, opts, err := connectDB(dbc, opts)
	cls := io.Closer(iokit.CloserChain{})
	if !fu.BoolOption(dontclose(false), opts) {
		cls = db
	}
	if err != nil {
		return lazy.ErrorSink(errstr.Wrapf(0, err,"database connection error: %v", err.Error()))
	}
	drv := fu.StrOption(Driver(""), opts)

	schema := fu.StrOption(Schema(""), opts)
	if schema != "" {
		switch drv {
		case "mysql":
			_, err = db.Exec("use " + schema)
		case "postgres":
			_, err = db.Exec("set search_path to " + schema)
		}
	}
	if err != nil {
		cls.Close()
		return lazy.ErrorSink(errstr.Wrapf(0, err, "query error: %s", err.Error()))
	}

	tx, err := db.Begin()
	if err != nil {
		cls.Close()
		return lazy.ErrorSink(errstr.Wrapf(0, err, "database begin transaction error: %s", err.Error()))
	}

	table := fu.StrOption(Table(""), opts)
	if table == "" {
		panic("there is no table")
	}
	if fu.Option(ErrorIfExists, opts).Interface().(IfExists_) == DropIfExists {
		_, err := tx.Exec(sqlDropQuery(table, opts...))
		if err != nil {
			cls.Close()
			return lazy.ErrorSink(errstr.Wrapf(0, err, "drop table error: %s", err.Error()))
		}
	}

	batchLen := fu.IntOption(Batch(1), opts)
	var stmt *sql.Stmt
	created := false
	batch := []interface{}{}
	names := []string{}
	pk := []bool{}

	return lazy.Sink(func(val interface{}, closeErr error) (err error) {
		if val == nil {
			if closeErr == nil {
				if len(batch) > 0 {
					if stmt, err = batchInsertStmt(tx, names, pk, len(batch)/len(names), table, opts); err == nil {
						if _, err = stmt.Exec(batch...); err == nil {
							cls = iokit.CloserChain{stmt, cls}
						}
					}
				}
				if err == nil {
					err = tx.Commit()
				}
			}
			return cls.Close()
		}

		switch r := val.(type) {
		case struct{}:
			// skip
		case *data.Row:
			if !created {
				names = make([]string, r.Width())
				pk = make([]bool, r.Width())
				//drv := fu.StrOption(Driver(""), opts)
				ns := make([]string,r.Width())
				for i := range ns {
					ns[i] = r.Factory.Name(i)
				}
				var dsx ColDesc
				dsx, err = Describe(ns, opts)
				if err != nil {
					return
				}
				describe := func(i int) (colType, colName string, isPk bool) {
					v := ns[i]
					colType, colName, isPk = dsx(v)
					if colType == "" {
						colType = sqlTypeOf(r.Data[i].Type(), drv)
					}
					return
				}
				for i := range names {
					_, names[i], pk[i] = describe(i)
				}
				_, err = tx.Exec(sqlCreateQuery(len(ns), table, describe, opts))
				if err != nil {
					return errstr.Wrapf(0, err, "create table error: %s", err.Error())
				}
				created = true
			}
			if len(batch)/len(names) >= batchLen {
				if stmt == nil {
					stmt, err = batchInsertStmt(tx, names, pk, len(batch)/len(names), table, opts)
					if err != nil {
						return err
					}
					cls = iokit.CloserChain{stmt, cls}
				}
				_, err = stmt.Exec(batch...)
				if err != nil {
					return err
				}
				batch = batch[:0]
			}
			for _, x := range r.Data {
				if x.Val == nil {
					batch = append(batch, nil)
				} else {
					batch = append(batch, x.Val)
				}
			}
		default:
			return errstr.Errorf("unsupported value type %v",fu.TypeOf(val))
		}
		return
	})
}

func sqlCreateQuery(n int, table string, describe func(int) (string, string, bool), opts []interface{}) string {
	pk := []string{}
	query := "create table "

	ifExists := fu.Option(ErrorIfExists, opts).Interface().(IfExists_)
	if ifExists != ErrorIfExists && ifExists != DropIfExists {
		query += "if not exists "
	}

	query = query + table + "( "
	for i := 0; i < n; i++ {
		if i != 0 {
			query += ", "
		}
		colType, colName, isPK := describe(i)
		query = query + colName + " " + colType
		if isPK {
			pk = append(pk, colName)
		}
	}

	if len(pk) > 0 {
		query = query + ", primary key (" + strings.Join(pk, ",") + ")"
	}

	query += " )"
	return query
}

func sqlDropQuery(table string, opts ...interface{}) string {
	schema := fu.StrOption(Schema(""), opts)
	if schema != "" {
		schema = schema + "."
	}
	return "drop table if exists " + schema + table
}

func sqlTypeOf(tp reflect.Type, driver string) string {
	switch tp.Kind() {
	case reflect.String:
		if driver == "postgres" {
			return "VARCHAR(65535)" /* redshift TEXT == VARCHAR(256) */
		}
		return "TEXT"
	case reflect.Int8, reflect.Uint8, reflect.Int16:
		return "SMALLINT"
	case reflect.Uint16, reflect.Int32, reflect.Int:
		return "INTEGER"
	case reflect.Uint, reflect.Uint32, reflect.Int64, reflect.Uint64:
		return "BIGINT"
	case reflect.Float32:
		if driver == "postgres" {
			return "REAL" /* redshift does not FLOAT */
		}
		return "FLOAT"
	case reflect.Float64:
		if driver == "postgres" {
			return "DOUBLE PRECISION" /* redshift does not have DOUBLE */
		}
		return "DOUBLE"
	case reflect.Bool:
		return "BOOLEAN"
	default:
		if tp == fu.Ts {
			return "DATETIME"
		}
	}
	panic("unsupported data type " + fmt.Sprintf("%v %v", tp.String(), tp.Kind()))
}
