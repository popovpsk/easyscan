package benchmarks

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgx/v4"
)

type pgxQuery interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

func sqlSelect(ctx context.Context, q pgxQuery, dest interface{}, query string, args ...interface{}) error {
	slicePtr := reflect.ValueOf(dest)

	//получен не nil ptr
	if slicePtr.Kind() != reflect.Ptr {
		return errors.New("must pass a pointer")
	}
	if slicePtr.IsNil() {
		return errors.New("nil pointer passed to Select dest")
	}

	//разыменовывание указателя
	slice := slicePtr.Elem()

	//получен слайс
	sliceType := slice.Type()
	if sliceType.Kind() != reflect.Slice {
		return fmt.Errorf("expected a slice but got %s", slice.Type().Kind())
	}

	//тип элемента слайса. Например это будет string, если dest = *[]string
	sliceElem := sliceType.Elem()

	//тип экземпляра
	exemplarType := deref(sliceElem)

	if exemplarType.Kind() != reflect.Struct && !isPrimitiveType(exemplarType.Kind()) {
		return fmt.Errorf("expected a struct or a pointer to a struct in the slice but got %s", exemplarType.Kind())
	}

	//isPtr - получен массив []*object или []object
	isPtr := sliceElem.Kind() == reflect.Ptr

	//запрос в бд
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	if isPrimitiveType(exemplarType.Kind()) || isPrimitiveStruct(exemplarType) {
		return scanPrimitives(rows, isPtr, slice, exemplarType)
	}

	return scanObjects(rows, isPtr, slice, exemplarType)
}

func scanPrimitives(rows pgx.Rows, isPtr bool, slice reflect.Value, exemplarType reflect.Type) error {
	for rows.Next() {
		exemplarPointer := reflect.New(exemplarType)

		err := rows.Scan(exemplarPointer.Elem().Addr().Interface())
		if err != nil {
			return fmt.Errorf("rows.Scan: %w", err)
		}

		addToSlice(slice, exemplarPointer, isPtr)
	}
	return rows.Err()
}

func scanObjects(rows pgx.Rows, isPtr bool, slice reflect.Value, exemplarType reflect.Type) error {
	//карта столбцов на номера полей класса
	fieldsNumbers := getFieldsNumbers(exemplarType, rows.FieldDescriptions())
	scans := make([]interface{}, len(fieldsNumbers))

	for rows.Next() {
		exemplarPointer := reflect.New(exemplarType)

		//scans - слайс указателей на поля структуры куда будем записывать результат
		//для структуры Person{id, name, age} создается слайс c указателями на поля экземпляра
		//для rows.Scan не отличимый от вызова rows.Scan(&p.id,&p.name,&p.age)
		fillScans(fieldsNumbers, scans, exemplarPointer)

		err := rows.Scan(scans...)
		if err != nil {
			return fmt.Errorf("rows.Scan: %w", err)
		}

		addToSlice(slice, exemplarPointer, isPtr)
	}

	return rows.Err()
}

func sqlGet(ctx context.Context, q pgxQuery, dest interface{}, query string, args ...interface{}) error {
	objectPtr := reflect.ValueOf(dest)

	//получили ptr не nil
	if objectPtr.Kind() != reflect.Ptr {
		return errors.New("must pass a pointer")
	}

	if objectPtr.IsNil() {
		return errors.New("nil pointer passed to Get dest")
	}

	//разыменовывание указателя
	objectType := objectPtr.Type().Elem()

	//по указателю - структура или примитив
	objectTypeKind := objectType.Kind()

	if objectTypeKind != reflect.Struct && !isPrimitiveType(objectTypeKind) {
		return fmt.Errorf("expected a struct or a primitive type but got %s", objectTypeKind)
	}

	//запрос в бд
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return pgx.ErrNoRows
	}

	//сканирование в поля полученной по указателю структуры либо если это примитив прям в него.
	if isPrimitiveType(objectTypeKind) || isPrimitiveStruct(objectType) {
		err = rows.Scan(objectPtr.Elem().Addr().Interface())
	} else {
		fieldsNumbers := getFieldsNumbers(objectType, rows.FieldDescriptions())
		scans := make([]interface{}, len(fieldsNumbers))
		fillScans(fieldsNumbers, scans, objectPtr)
		err = rows.Scan(scans...)
	}

	if err != nil {
		return fmt.Errorf("rows.Scan: %w", err)
	}

	if rows.Next() {
		return errors.New("sql: expected one row")
	}

	return rows.Err()
}

const undefinedField = -1

// getFieldsNumbers строит список номеров полей по порядковому номеру столбцов. -1 если в структуре нет нужного поля.
func getFieldsNumbers(t reflect.Type, fieldDescriptions []pgproto3.FieldDescription) []int {
	result := make([]int, len(fieldDescriptions))
	for i := range result {
		result[i] = undefinedField
	}

	tags := getStructDBTags(t)

iterateFields:
	for idx, fd := range fieldDescriptions {
		for i := 0; i < len(tags); i++ {
			if string(fd.Name) == tags[i] {
				result[idx] = i
				continue iterateFields
			}
		}
	}
	return result
}

func fillScans(fieldsNumbers []int, scans []interface{}, exemplarPointer reflect.Value) {
	for i, fieldNumber := range fieldsNumbers {
		if fieldNumber == undefinedField {
			scans[i] = emptyScanObj
		} else {
			scans[i] = exemplarPointer.Elem().Field(fieldNumber).Addr().Interface()
		}
	}
}

var tagsCache = make(map[reflect.Type][]string)
var tagsCacheL sync.RWMutex

func getStructDBTags(t reflect.Type) []string {
	tagsCacheL.RLock()
	tags, ok := tagsCache[t]
	tagsCacheL.RUnlock()

	if ok {
		return tags
	}

	numField := t.NumField()
	result := make([]string, 0, numField)
	for i := 0; i < numField; i++ {
		result = append(result, t.Field(i).Tag.Get("db"))
	}

	tagsCacheL.Lock()
	defer tagsCacheL.Unlock()

	tagsCache[t] = result

	return result
}

var emptyScanObj emptyScan

type emptyScan struct {
}

func (emptyScan) Scan(_ interface{}) error {
	return nil
}

func deref(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// addToSlice добавляет element в конец slice
func addToSlice(slice reflect.Value, element reflect.Value, isPtr bool) {
	// append
	l := slice.Len()
	if l < slice.Cap() {
		l++
		slice.SetLen(l)

		if isPtr {
			slice.Index(l - 1).Set(element)
		} else {
			slice.Index(l - 1).Set(reflect.Indirect(element))
		}
		return
	}

	if isPtr {
		slice.Set(reflect.Append(slice, element))
	} else {
		slice.Set(reflect.Append(slice, reflect.Indirect(element)))
	}
}

var timeType = reflect.TypeOf(time.Time{})

func isPrimitiveStruct(objectType reflect.Type) bool {
	if objectType.Kind() == reflect.Struct {
		switch objectType {
		case timeType:
			return true
		}
	}
	return false
}

func isPrimitiveType(kind reflect.Kind) bool {
	switch kind {
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
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,

		reflect.String:
		return true
	default:
		return false
	}
}
