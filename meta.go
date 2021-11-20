package data

import "reflect"

type Meta interface {
	Type() reflect.Type
	Convert(string, *interface{}, int, int) error
	Format(interface{}) (string,error)
}
