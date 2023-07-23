package easyscan

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"
	"time"
)

type customScanner struct {
}

func (cs *customScanner) Scan(_ interface{}) error {
	return nil
}

func Test_isSupportedType(t *testing.T) {
	type alias = sql.NullString

	tests := []struct {
		name string
		arg  reflect.Type
		want bool
	}{
		{
			name: "int",
			arg:  reflect.TypeOf(1),
			want: true,
		},
		{
			name: "string",
			arg:  reflect.TypeOf(""),
			want: true,
		},
		{
			name: "slice",
			arg:  reflect.TypeOf([]int{1, 2, 3}),
			want: true,
		},
		{
			name: "array",
			arg:  reflect.TypeOf([3]int{1, 2, 3}),
			want: true,
		},
		{
			name: "time",
			arg:  reflect.TypeOf(time.Now()),
			want: true,
		},
		{
			name: "nullString",
			arg:  reflect.TypeOf(sql.NullString{}),
			want: true,
		},
		{
			name: "duration",
			arg:  reflect.TypeOf(time.Duration(0)),
			want: true,
		},
		{
			name: "alias type",
			arg:  reflect.TypeOf(alias{}),
			want: true,
		},
		{
			name: "custom scanner",
			arg:  reflect.TypeOf(customScanner{}),
			want: true,
		},
		{
			name: "isn't scanner",
			arg:  reflect.TypeOf(strings.Builder{}),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSupportedType(tt.arg); got != tt.want {
				t.Errorf("isSupportedType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_implementsScanner(t *testing.T) {
	equal(t, true, implementsScanner(reflect.TypeOf(sql.NullTime{})))
	equal(t, false, implementsScanner(reflect.TypeOf(time.Time{})))
}
