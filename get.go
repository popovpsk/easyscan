package easyscan

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/jackc/pgx/v4"
)

var ErrMoreThanOneRow = errors.New("get expects 1 row")

func Get(ctx context.Context, conn pgxExecutor, dest interface{}, query string, args ...interface{}) error {
	if conn == nil {
		return errors.New("conn is nil")
	}

	objectPtr := reflect.ValueOf(dest)
	if objectPtr.Kind() != reflect.Ptr || objectPtr.IsNil() {
		return errors.New("destination must be a non nil pointer")
	}

	objectType := objectPtr.Type().Elem()

	isPgxSupported := isPgxSupportedType(objectType)
	if objectType.Kind() != reflect.Struct && !isPgxSupported {
		return fmt.Errorf("expected a struct or a pgx supported type but got %s", objectType.Kind())
	}

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return pgx.ErrNoRows
	}

	if isPgxSupported {
		err = rows.Scan(dest)
	} else {
		scans, e := getSliceForScan(objectType, rows.FieldDescriptions(), objectPtr)
		if e != nil {
			return e
		}
		err = rows.Scan(scans...)
	}

	if err != nil {
		return fmt.Errorf("pgx.rows.Scan: %w", err)
	}

	if rows.Next() {
		return ErrMoreThanOneRow
	}

	return rows.Err()
}
