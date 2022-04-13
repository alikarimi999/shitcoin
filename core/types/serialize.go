package types

import (
	"reflect"
)

var serialType = reflect.TypeOf((*serializer)(nil)).Elem()

type serializer interface {
	serialize() []byte
}

func join(i interface{}) []byte {

	result := []byte{}
	v := reflect.ValueOf(i)
	t := v.Type()
	if t.Kind() == reflect.Slice && t.Elem().Implements(serialType) {
		switch t.Elem().String() {
		case "*types.Block":
			for i := 0; i < v.Len(); i++ {
				result = append(result, v.Index(i).Interface().(*Block).serialize()...)
			}
		case "*types.Transaction":
			for i := 0; i < v.Len(); i++ {
				result = append(result, v.Index(i).Interface().(*Transaction).serialize()...)
			}
		case "*types.TxIn":
			for i := 0; i < v.Len(); i++ {
				result = append(result, v.Index(i).Interface().(*TxIn).serialize()...)
			}
		case "*types.TxOut":
			for i := 0; i < v.Len(); i++ {
				result = append(result, v.Index(i).Interface().(*TxOut).serialize()...)
			}
		}
	}

	return result
}
