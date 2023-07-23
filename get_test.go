package easyscan

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func TestGet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, connString)
	noError(t, err)
	noError(t, err)
	defer pool.Close()

	t.Run("primitive", func(t *testing.T) {
		var result int
		err = Get(ctx, pool, &result, "SELECT 1")
		noError(t, err)
		equal(t, 1, result)
	})

	t.Run("string", func(t *testing.T) {
		var result string
		err = Get(ctx, pool, &result, "SELECT 'foo'")
		noError(t, err)
		equal(t, "foo", result)
	})

	t.Run("bool", func(t *testing.T) {
		var result bool
		err = Get(ctx, pool, &result, "SELECT 1=1")
		noError(t, err)
		equal(t, true, result)
	})

	t.Run("time", func(t *testing.T) {
		var result time.Time
		err = Get(ctx, pool, &result, "SELECT '2012-03-04 10:11:12'::timestamp")
		noError(t, err)
		equal(t, time.Date(2012, 3, 4, 10, 11, 12, 0, time.UTC), result.UTC())
	})

	t.Run("duration", func(t *testing.T) {
		var result time.Duration
		err = Get(ctx, pool, &result, "SELECT '1 hour'::interval")
		noError(t, err)
		equal(t, time.Hour, result)
	})

	t.Run("named type", func(t *testing.T) {
		type namedType int
		var result namedType
		err = Get(ctx, pool, &result, "SELECT 1")
		noError(t, err)
		equal(t, 1, int(result))
	})

	t.Run("string slice", func(t *testing.T) {
		var result []string
		err = Get(ctx, pool, &result, "SELECT ARRAY['1','2','3']")
		noError(t, err)
		equal(t, "1", result[0])
		equal(t, "2", result[1])
		equal(t, "3", result[2])
	})

	t.Run("array of rune slices", func(t *testing.T) {
		var result [3][]byte
		err = Get(ctx, pool, &result, "SELECT ARRAY['1','2','34']")
		noError(t, err)
		equal(t, byte('1'), result[0][0])
		equal(t, byte('2'), result[1][0])
		equal(t, byte('3'), result[2][0])
		equal(t, byte('4'), result[2][1])
	})

	t.Run("no rows", func(t *testing.T) {
		var result bool
		err = Get(ctx, pool, &result, "SELECT * FROM information_schema.tables WHERE 1=0")
		notNilError(t, err)

		if !errors.Is(err, pgx.ErrNoRows) {
			fmt.Printf("error %v is not pgs.ErrNoRows\n", err.Error())
			t.Fail()
		}
	})

	t.Run("not a pointer", func(t *testing.T) {
		var result int
		err = Get(ctx, pool, result, "SELECT 1")
		notNilError(t, err)
		errorContains(t, err, "must pass a pointer")
	})

	t.Run("nil ptr", func(t *testing.T) {
		var result *int
		err = Get(ctx, pool, result, "SELECT 1")
		notNilError(t, err)
		errorContains(t, err, "nil pointer passed")
	})

	t.Run("nil conn", func(t *testing.T) {
		var result *int
		err = Get(ctx, nil, result, "SELECT 1")
		notNilError(t, err)
		errorContains(t, err, "conn is nil")
	})
}

func TestGetStruct(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, connString)
	noError(t, err)
	defer pool.Close()

	const createTable = `
    CREATE TABLE if not exists easy_scan(
    id bigserial primary key,
    str1 text,
    ts1 timestamp,
    b1 bool)
`

	_, err = pool.Exec(ctx, createTable)
	noError(t, err)

	defer func() {
		_, e := pool.Exec(ctx, "DROP TABLE easy_scan")
		if e != nil {
			panic(e)
		}
	}()

	now := time.Now().Round(time.Minute).UTC()

	t.Run("struct", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			tag, err := pool.Exec(ctx, "INSERT INTO easy_scan(str1, ts1, b1) VALUES ($1, $2, $3)",
				fmt.Sprintf("%d", i),
				now,
				true)
			noError(t, err)
			equal(t, true, tag.Insert())
			equal(t, int64(1), tag.RowsAffected())
		}
		defer func() {
			_, e := pool.Exec(ctx, "TRUNCATE easy_scan")
			if e != nil {
				panic(e)
			}
		}()

		type easyScan struct {
			Id   int64     `db:"id"`
			Str1 string    `db:"str1"`
			Ts1  time.Time `db:"ts1"`
			B1   bool      `db:"b1"`
		}

		t.Run("value", func(t *testing.T) {
			result := new(easyScan)
			err = Get(ctx, pool, result, "SELECT * FROM easy_scan LIMIT 1")
			noError(t, err)

			equal(t, int64(1), result.Id)
			equal(t, "0", result.Str1)
			equal(t, now, result.Ts1)
			equal(t, true, result.B1)
		})

		t.Run("anonymous", func(t *testing.T) {
			result := new(struct {
				Id  int64  `db:"id"`
				Str string `db:"str1"`
			})
			err = Get(ctx, pool, result, "SELECT * FROM easy_scan LIMIT 1")
			noError(t, err)

			equal(t, int64(1), result.Id)
			equal(t, "0", result.Str)
		})

		t.Run("extra column", func(t *testing.T) {
			result := new(easyScan)
			err = Get(ctx, pool, result, "SELECT *, 1 as extra FROM easy_scan LIMIT 1")
			noError(t, err)

			equal(t, int64(1), result.Id)
			equal(t, "0", result.Str1)
			equal(t, now, result.Ts1)
			equal(t, true, result.B1)
		})

		t.Run("missing column", func(t *testing.T) {
			result := new(easyScan)
			err = Get(ctx, pool, result, "SELECT id, str1 FROM easy_scan LIMIT 1")
			noError(t, err)

			equal(t, int64(1), result.Id)
			equal(t, "0", result.Str1)
		})

		t.Run("with table alias", func(t *testing.T) {
			result := new(easyScan)
			err = Get(ctx, pool, result, "SELECT es.* FROM easy_scan AS es LIMIT 1")
			noError(t, err)

			equal(t, int64(1), result.Id)
			equal(t, "0", result.Str1)
			equal(t, now, result.Ts1)
			equal(t, true, result.B1)
		})
	})

	t.Run("nil pointers", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			tag, err := pool.Exec(ctx, "INSERT INTO easy_scan(str1, ts1, b1) VALUES ($1, $2, $3)",
				fmt.Sprintf("%d", i),
				nil,
				nil)
			noError(t, err)
			equal(t, true, tag.Insert())
			equal(t, int64(1), tag.RowsAffected())
		}
		defer func() {
			_, e := pool.Exec(ctx, "TRUNCATE easy_scan")
			if e != nil {
				panic(e)
			}
		}()

		type easyScan struct {
			Id   int64      `db:"id"`
			Str1 *string    `db:"str1"`
			Ts1  *time.Time `db:"ts1"`
			B1   *bool      `db:"b1"`
		}

		result := new(easyScan)
		err = Get(ctx, pool, result, "SELECT * FROM easy_scan LIMIT 1")
		noError(t, err)

		var nilTime *time.Time
		var nilBool *bool
		equal(t, "0", *result.Str1)
		equal(t, nilTime, result.Ts1)
		equal(t, nilBool, result.B1)
	})
}

func TestGetErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, connString)
	noError(t, err)
	defer pool.Close()

	t.Run("no rows", func(t *testing.T) {
		var result int
		const query = "select 1 where 1=0"
		pgxErr := pool.QueryRow(ctx, query).Scan(&result)

		easyscanErr := Get(ctx, pool, &result, query)
		equal(t, pgxErr, easyscanErr)
		equal(t, pgxErr, pgx.ErrNoRows)
	})

	t.Run("more than 1 rows", func(t *testing.T) {
		var result int
		const query = "SELECT generate_series(0, 9)"

		easyscanErr := Get(ctx, pool, &result, query)
		equal(t, ErrMoreThanOneRow, easyscanErr)
	})
}
