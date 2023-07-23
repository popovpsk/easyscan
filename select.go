package easyscan

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgx/v4"
)

var emptyScanObj = emptyScan{}

type emptyScan struct {
}

func (emptyScan) Scan(_ interface{}) error {
	return nil
}

type pgxExecutor interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

func Select(ctx context.Context, conn pgxExecutor, dest interface{}, query string, args ...interface{}) error {
	if conn == nil {
		return errors.New("conn is nil")
	}

	slicePtr := reflect.ValueOf(dest)

	if slicePtr.Kind() != reflect.Ptr {
		return errors.New("must pass a pointer")
	}
	if slicePtr.IsNil() {
		return errors.New("nil pointer passed to Select dest")
	}

	slice := slicePtr.Elem()

	sliceType := slice.Type()
	if sliceType.Kind() != reflect.Slice {
		return fmt.Errorf("expected a slice but got %s", slice.Type().Kind())
	}

	//example: is string, for dest = *[]string
	sliceElemType := sliceType.Elem()

	//[]*object or []object
	isPtr := sliceElemType.Kind() == reflect.Ptr

	exemplarType := sliceElemType
	if isPtr {
		exemplarType = exemplarType.Elem()
	}

	if exemplarType.Kind() != reflect.Struct && !isSupportedType(exemplarType) {
		return fmt.Errorf("expected a struct or a pointer to a struct in the slice but got %s", exemplarType.Kind())
	}

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	if isSupportedType(exemplarType) {
		return scanToSupported(rows, isPtr, slice, exemplarType)
	}

	return scanObjects(rows, isPtr, slice, exemplarType)
}

func scanToSupported(rows pgx.Rows, isPtr bool, slice reflect.Value, exemplarType reflect.Type) error {
	for rows.Next() {
		exemplarPointer := reflect.New(exemplarType)

		err := rows.Scan(exemplarPointer.Interface())
		if err != nil {
			return fmt.Errorf("rows.Scan: %w", err)
		}

		if isPtr {
			addToSlice(slice, exemplarPointer)
		} else {
			addToSlice(slice, exemplarPointer.Elem())
		}
	}
	return rows.Err()
}

func scanObjects(rows pgx.Rows, isPtr bool, slice reflect.Value, exemplarType reflect.Type) error {
	if !rows.Next() {
		return rows.Err()
	}

	objectForFilling := reflect.New(exemplarType)
	scans, err := getSliceForScan(exemplarType, rows.FieldDescriptions(), objectForFilling)
	if err != nil {
		return err
	}

	err = rows.Scan(scans...)
	if err != nil {
		return fmt.Errorf("rows.Scan: %w", err)
	}

	for rows.Next() {
		if isPtr {
			exemplarPtr := reflect.New(exemplarType)
			exemplarPtr.Elem().Set(objectForFilling.Elem())
			addToSlice(slice, exemplarPtr)
		} else {
			addToSlice(slice, objectForFilling.Elem())
		}

		err = rows.Scan(scans...)
		if err != nil {
			return fmt.Errorf("rows.Scan: %w", err)
		}
	}

	if isPtr {
		addToSlice(slice, objectForFilling)
	} else {
		addToSlice(slice, objectForFilling.Elem())
	}

	return rows.Err()
}

func addToSlice(slice reflect.Value, element reflect.Value) {
	l := slice.Len()
	if l < slice.Cap() {
		l++
		slice.SetLen(l)
		slice.Index(l - 1).Set(element)
		return
	}

	slice.Set(reflect.Append(slice, element))
}

func getSliceForScan(t reflect.Type, fieldDescriptions []pgproto3.FieldDescription, exemplarPointer reflect.Value) ([]interface{}, error) {
	scans := make([]interface{}, len(fieldDescriptions))

	tags := getStructDBTags(t)

	e := exemplarPointer.Elem()

	matchingFailed := true

	for idx, fd := range fieldDescriptions {
		fieldIdx, ok := tags.find(fd.Name)
		if ok {
			matchingFailed = false
			scans[idx] = e.Field(fieldIdx).Addr().Interface()
		} else {
			scans[idx] = emptyScanObj
		}
	}

	if matchingFailed {
		return nil, errors.New("db tags have no matches to columns")
	}

	return scans, nil
}
