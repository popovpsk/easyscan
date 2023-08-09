package easyscan

import (
	"database/sql"
	"reflect"
	"time"
)

var timeType = reflect.TypeOf(time.Time{})
var scannerType = reflect.TypeOf((*sql.Scanner)(nil)).Elem()

func implementsScanner(objectType reflect.Type) bool {
	return reflect.PtrTo(objectType).Implements(scannerType)
}

func isPgxSupportedType(objectType reflect.Type, first bool) bool {
	switch objectType.Kind() {
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64,
		reflect.String,
		reflect.Array,
		reflect.Slice:
		return true

	case reflect.Struct:
		switch objectType {
		case timeType:
			return true
		}

		if implementsScanner(objectType) {
			return true
		}
	case reflect.Ptr:
		return first && isPgxSupportedType(objectType.Elem(), false)
	}
	return false
}
