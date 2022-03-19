package data

import (
	"fmt"
	"go4ml.xyz/errstr"
	"go4ml.xyz/fu"
	"reflect"
	"sync"
)

var enumType = reflect.TypeOf(Enum{})

/*
Enum encapsulate enumeration abstraction in relation to tables
*/
type Enum struct {
	Text  string
	Value int
}

// Text return enum string representation
func (e Enum) String() string {
	return e.Text
}

// Enum defines enumerated meta-column with the Enum type
func (e Enumset) Enum() Meta {
	return Enumerator{e, &sync.Mutex{}, len(e) != 0}
}

// Enum defines enumerated meta-column with the string type
func (e Enumset) Text() Meta {
	return TextEnumerator{Enumerator{e, &sync.Mutex{}, len(e) != 0}}
}

// Enum defines enumerated meta-column with the int type
func (e Enumset) Integer() Meta {
	return IntegerEnumerator{
		Enumerator{e, &sync.Mutex{}, len(e) != 0},
		fu.KeysOf((map[string]int)(e)).([]string),
	}
}

// Enumset is a set of values belongs to one enumeration
type Enumset map[string]int

// Len returns length of enumset aka count of enum values
func (m Enumset) Len() int {
	return len(m)
}

// Enumerator the object enumerates enums in data stream
type Enumerator struct {
	m  Enumset
	mu *sync.Mutex
	ro bool
}

func (ce Enumerator) enumerate(v string) (e int, ok bool, err error) {
	ce.mu.Lock()
	if e, ok = ce.m[v]; !ok {
		if ce.ro {
			return 0, false, errstr.Errorf("readonly enumset does not have value `%v`" + v)
		}
		ce.m[v] = len(ce.m)
	}
	ce.mu.Unlock()
	return
}

// Type returns the type of column
func (ce Enumerator) Type() reflect.Type {
	return enumType // it's the Enum meta-column
}
func (ce Enumerator) Convert(v string, value *interface{}, _, _ int) error {
	if v != "" {
		e, _, err := ce.enumerate(v)
		if err != nil {
			return err
		}
		*value = Enum{v, e}
	}
	return nil
}
func (ce Enumerator) Format(x interface{}) (string, error) {
	if x == nil { // format N/A value
		return "", nil
	}
	if fu.TypeOf(x) == enumType {
		text := x.(Enum).Text
		if _, ok := ce.m[text]; ok {
			return text, nil
		}
	}
	return "", errstr.Errorf("`%v` is not an enumeration value", x)
}

type IntegerEnumerator struct {
	Enumerator
	rev []string
}

func (ce IntegerEnumerator) Type() reflect.Type {
	return fu.Int
}

func (ce IntegerEnumerator) Convert(v string, value *interface{}, _, _ int) error {
	if v != "" {
		e, ok, err := ce.enumerate(v)
		if err != nil {
			return err
		}
		if !ok {
			ce.mu.Lock()
			ce.rev = append(ce.rev, v)
			ce.mu.Unlock()
		}
		*value = e
	}
	return nil
}

func (ce IntegerEnumerator) Format(v interface{}) (string, error) {
	if v == nil { // format N/A value
		return "", nil
	}
	if text, ok := v.(string); ok {
		if e, ok := ce.m[text]; ok {
			return fmt.Sprint(e), nil
		}
	}
	return "", errstr.Errorf("`%v` is not an enumeration value", v)
}

type TextEnumerator struct{ Enumerator }

func (ce TextEnumerator) Type() reflect.Type {
	return fu.String
}

func (ce TextEnumerator) Convert(v string, value *interface{}, _, _ int) error {
	if v != "" {
		_, _, err := ce.enumerate(v)
		if err != nil {
			return err
		}
		*value = v
	}
	return nil
}
