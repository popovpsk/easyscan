package easyscan

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

const connString = "user=postgres password=postgres host=localhost dbname=easyscan port=5432"

func TestSelect(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, connString)
	noError(t, err)
	defer pool.Close()

	t.Run("primitive", func(t *testing.T) {
		var result []int
		err = Select(ctx, pool, &result, "SELECT generate_series(0, 9)")
		noError(t, err)
		if len(result) != 10 {
			t.Fail()
			panic("len is not 10")
		}
		for i, v := range result {
			equal(t, i, v)
		}
	})

	t.Run("ptr to primitive", func(t *testing.T) {
		var result []*int
		err = Select(ctx, pool, &result, "SELECT generate_series(0, 9)")
		noError(t, err)
		if len(result) != 10 {
			t.Fail()
			panic("len is not 10")
		}
		for i, v := range result {
			equal(t, i, *v)
		}
	})

	t.Run("named types", func(t *testing.T) {
		type namedType string
		var result []namedType
		err = Select(ctx, pool, &result, "SELECT * FROM UNNEST(ARRAY['foo', 'bar'])")
		noError(t, err)
		if len(result) != 2 {
			t.Fail()
			panic("len is not 2")
		}
		equal(t, "foo", string(result[0]))
		equal(t, "bar", string(result[1]))
	})

	t.Run("string", func(t *testing.T) {
		var result []string
		err = Select(ctx, pool, &result, "SELECT * FROM UNNEST(ARRAY['foo', 'bar'])")
		noError(t, err)
		if len(result) != 2 {
			t.Fail()
			panic("len is not 2")
		}
		equal(t, "foo", result[0])
		equal(t, "bar", result[1])
	})

	t.Run("time", func(t *testing.T) {
		var result []time.Time
		err = Select(ctx, pool, &result, "SELECT * FROM UNNEST(ARRAY['2012-03-04 10:11:12'::timestamp, '2011-11-11 12:11:10'::timestamp])")
		noError(t, err)
		equal(t, time.Date(2012, 3, 4, 10, 11, 12, 0, time.UTC), result[0])
		equal(t, time.Date(2011, 11, 11, 12, 11, 10, 0, time.UTC), result[1])
	})

	t.Run("named time", func(t *testing.T) {
		t.Skip()
		type namedType time.Time
		var result []namedType
		err = Select(ctx, pool, &result, "SELECT * FROM UNNEST(ARRAY['2012-03-04 10:11:12'::timestamp, '2011-11-11 12:11:10'::timestamp])")
		noError(t, err)
		equal(t, time.Date(2012, 3, 4, 10, 11, 12, 0, time.UTC), time.Time(result[0]))
		equal(t, time.Date(2011, 11, 11, 12, 11, 10, 0, time.UTC), time.Time(result[1]))
	})

	t.Run("slice int", func(t *testing.T) {
		var result [][]int
		err = Select(ctx, pool, &result, "SELECT ARRAY[1,2,3] UNION ALL SELECT ARRAY[4,5,6]")
		noError(t, err)
		equal(t, 1, result[0][0])
		equal(t, 2, result[0][1])
		equal(t, 3, result[0][2])
		equal(t, 4, result[1][0])
		equal(t, 5, result[1][1])
		equal(t, 6, result[1][2])
	})

	t.Run("array byte", func(t *testing.T) {
		var result [][3]byte
		err = Select(ctx, pool, &result, "SELECT ARRAY[1,2,3] UNION ALL SELECT ARRAY[4,5,6]")
		noError(t, err)
		equal(t, byte(1), result[0][0])
		equal(t, byte(2), result[0][1])
		equal(t, byte(3), result[0][2])
		equal(t, byte(4), result[1][0])
		equal(t, byte(5), result[1][1])
		equal(t, byte(6), result[1][2])
	})

	t.Run("no rows", func(t *testing.T) {
		var result []int
		err = Select(ctx, pool, &result, "SELECT * FROM information_schema.tables WHERE 1=0")
		noError(t, err)
		if len(result) != 0 {
			t.Fail()
			panic("len is not 0")
		}
	})

	t.Run("not a pointer", func(t *testing.T) {
		var result []int
		err = Select(ctx, pool, result, "SELECT 1")
		notNilError(t, err)
		errorContains(t, err, "must be a non nil pointer to slice")
	})

	t.Run("not a slice", func(t *testing.T) {
		var result int
		err = Select(ctx, pool, &result, "SELECT 1")
		notNilError(t, err)
		errorContains(t, err, "expected a slice but got")
	})

	t.Run("does not have required db tags ", func(t *testing.T) {
		type t1 struct {
			ID int `db:"id"`
		}
		result := make([]t1, 0)
		err = Select(ctx, pool, &result, "SELECT * FROM UNNEST(ARRAY['foo', 'bar']) AS name")
		notNilError(t, err)
		errorContains(t, err, "have no matches to columns")
	})

	t.Run("nil slice", func(t *testing.T) {
		var result *[]string
		err = Select(ctx, pool, result, "SELECT * FROM UNNEST(ARRAY['foo', 'bar'])")
		notNilError(t, err)
		errorContains(t, err, "must be a non nil pointer to slice")
	})

	t.Run("nil conn", func(t *testing.T) {
		var result *[]string
		err = Select(ctx, nil, result, "SELECT * FROM UNNEST(ARRAY['foo', 'bar'])")
		notNilError(t, err)
		errorContains(t, err, "conn is nil")
	})
}

func TestSelectStruct(t *testing.T) {
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

		t.Run("values", func(t *testing.T) {
			result := make([]easyScan, 0, 5)
			err = Select(ctx, pool, &result, "SELECT * FROM easy_scan")
			noError(t, err)
			if len(result) != 5 {
				t.Fail()
				panic("len is not 5")
			}
			equal(t, 5, cap(result))
			for i := 0; i < 5; i++ {
				equal(t, int64(i+1), result[i].Id)
				equal(t, fmt.Sprintf("%d", i), result[i].Str1)
				equal(t, now, result[i].Ts1)
				equal(t, true, result[i].B1)
			}
		})

		t.Run("pointers", func(t *testing.T) {
			result := make([]*easyScan, 0)
			err = Select(ctx, pool, &result, "SELECT * FROM easy_scan")
			noError(t, err)
			if len(result) != 5 {
				t.Fail()
				panic("len is not 5")
			}
			for i := 0; i < 5; i++ {
				equal(t, int64(i+1), result[i].Id)
				equal(t, fmt.Sprintf("%d", i), result[i].Str1)
				equal(t, now, result[i].Ts1)
				equal(t, true, result[i].B1)
			}
		})

		t.Run("empty values", func(t *testing.T) {
			result := make([]easyScan, 0)
			err = Select(ctx, pool, &result, "SELECT * FROM easy_scan WHERE 1=0")
			noError(t, err)
			if len(result) != 0 {
				t.Fail()
				panic("len is not 0")
			}
		})

		t.Run("empty pointers", func(t *testing.T) {
			result := make([]*easyScan, 0)
			err = Select(ctx, pool, &result, "SELECT * FROM easy_scan WHERE 1=0")
			noError(t, err)
			if len(result) != 0 {
				t.Fail()
				panic("len is not 0")
			}
		})

		t.Run("anonymous", func(t *testing.T) {
			result := make([]struct {
				Id  int64  `db:"id"`
				Str string `db:"str1"`
			}, 0)
			err = Select(ctx, pool, &result, "SELECT * FROM easy_scan")
			noError(t, err)
			if len(result) != 5 {
				t.Fail()
				panic("len is not 5")
			}
			for i := 0; i < 5; i++ {
				equal(t, int64(i+1), result[i].Id)
				equal(t, fmt.Sprintf("%d", i), result[i].Str)
			}
		})

		t.Run("extra column", func(t *testing.T) {
			result := make([]easyScan, 0, 5)
			err = Select(ctx, pool, &result, "SELECT *, 1 as extra FROM easy_scan")
			noError(t, err)
			if len(result) != 5 {
				t.Fail()
				panic("len is not 5")
			}
			equal(t, 5, cap(result))
			for i := 0; i < 5; i++ {
				equal(t, int64(i+1), result[i].Id)
				equal(t, fmt.Sprintf("%d", i), result[i].Str1)
				equal(t, now, result[i].Ts1)
				equal(t, true, result[i].B1)
			}
		})

		t.Run("missing column", func(t *testing.T) {
			result := make([]easyScan, 0, 5)
			err = Select(ctx, pool, &result, "SELECT id, str1 FROM easy_scan")
			noError(t, err)
			if len(result) != 5 {
				t.Fail()
				panic("len is not 5")
			}
			equal(t, 5, cap(result))
			for i := 0; i < 5; i++ {
				equal(t, int64(i+1), result[i].Id)
				equal(t, fmt.Sprintf("%d", i), result[i].Str1)
			}
		})

		t.Run("with table alias", func(t *testing.T) {
			result := make([]easyScan, 0, 5)
			err = Select(ctx, pool, &result, "SELECT es.* FROM easy_scan AS es")
			noError(t, err)
			if len(result) != 5 {
				t.Fail()
				panic("len is not 5")
			}
			equal(t, 5, cap(result))
			for i := 0; i < 5; i++ {
				equal(t, int64(i+1), result[i].Id)
				equal(t, fmt.Sprintf("%d", i), result[i].Str1)
				equal(t, now, result[i].Ts1)
				equal(t, true, result[i].B1)
			}
		})
	})

	t.Run("scanner (ptr)", func(t *testing.T) {
		result := make([]sql.NullString, 0)
		err = Select(ctx, pool, &result, "SELECT * FROM UNNEST(ARRAY['foo', NULL])")
		noError(t, err)
		if len(result) != 2 {
			t.Fail()
			panic("len is not 2")
		}
		equal(t, "foo", result[0].String)
		equal(t, true, result[0].Valid)
		equal(t, "", result[1].String)
		equal(t, false, result[1].Valid)
	})

	t.Run("unsupported types", func(t *testing.T) {
		result := make([]chan int, 0)
		err = Select(ctx, pool, &result, "SELECT * FROM UNNEST(ARRAY['foo', NULL])")
		errorContains(t, err, "expected a struct")

		result2 := make([]*map[int64]string, 0)
		err = Select(ctx, pool, &result2, "SELECT * FROM UNNEST(ARRAY['foo', NULL])")
		errorContains(t, err, "expected a struct")
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

		result := make([]easyScan, 0, 5)
		err = Select(ctx, pool, &result, "SELECT * FROM easy_scan")
		noError(t, err)
		if len(result) != 5 {
			t.Fail()
			panic("len is not 5")
		}
		equal(t, 5, cap(result))
		var nilTime *time.Time
		var nilBool *bool
		for i := 0; i < 5; i++ {
			equal(t, fmt.Sprintf("%d", i), *result[i].Str1)
			equal(t, nilTime, result[i].Ts1)
			equal(t, nilBool, result[i].B1)
		}
	})
}

func TestConcurrency(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, connString)
	noError(t, err)
	defer pool.Close()

	const createTable = `
    CREATE TABLE if not exists easy_scan(
    id bigserial primary key,
    str1 text,
    ts1 timestamp,
    b1 bool)`

	_, err = pool.Exec(ctx, createTable)
	noError(t, err)

	defer func() {
		_, e := pool.Exec(ctx, "DROP TABLE easy_scan")
		if e != nil {
			panic(e)
		}
	}()

	const insertSql = `
INSERT INTO easy_scan(str1, ts1, b1)
VALUES ('1', '2012-03-04 10:11:12', true),
        (null, null, null),
        ('2', '2012-03-04 10:11:13', true),
        ('3', '2012-03-04 10:11:14', false)`

	tag, err := pool.Exec(ctx, insertSql)
	noError(t, err)
	equal(t, int64(4), tag.RowsAffected())

	type easyScan struct {
		Id   int64      `db:"id"`
		Str1 *string    `db:"str1"`
		Ts1  *time.Time `db:"ts1"`
		B1   *bool      `db:"b1"`
	}

	const iterations = 1000
	sema := make(chan struct{}, 64)
	wg := new(sync.WaitGroup)
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		sema <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sema }()

			var err error
			resultValues := make([]easyScan, 0)
			err = Select(ctx, pool, &resultValues, "SELECT * FROM easy_scan")
			noError(t, err)

			resultPointers := make([]*easyScan, 0)
			err = Select(ctx, pool, &resultPointers, "SELECT * FROM easy_scan")
			noError(t, err)

			m := &easyScan{Id: 1, Str1: stringPtr("1"), B1: boolPtr(true), Ts1: timePtr(time.Date(2012, 3, 4, 10, 11, 12, 0, time.UTC))}
			equal(t, m, &resultValues[0])
			equal(t, m, resultPointers[0])

			m = &easyScan{Id: 2}
			equal(t, m, &resultValues[1])
			equal(t, m, resultPointers[1])

			m = &easyScan{Id: 3, Str1: stringPtr("2"), B1: boolPtr(true), Ts1: timePtr(time.Date(2012, 3, 4, 10, 11, 13, 0, time.UTC))}
			equal(t, m, &resultValues[2])
			equal(t, m, resultPointers[2])

			m = &easyScan{Id: 4, Str1: stringPtr("3"), B1: boolPtr(false), Ts1: timePtr(time.Date(2012, 3, 4, 10, 11, 14, 0, time.UTC))}
			equal(t, m, &resultValues[3])
			equal(t, m, resultPointers[3])

			result := new(easyScan)
			err = Get(ctx, pool, result, "SELECT * FROM easy_scan WHERE id = 4")
			noError(t, err)
			equal(t, m, result)
		}()
	}

	wg.Wait()
}

func TestSelectEmbeddedTypes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, connString)
	noError(t, err)
	defer pool.Close()

	type RowChanges struct {
		CreatedAt time.Time  `db:"created_at"`
		UpdatedAt *time.Time `db:"updated_at"`
		DeletedAt *time.Time `db:"deleted_at"`
	}

	type RowInfo struct {
		RowChanges
		Version int `db:"version"`
	}

	type Person struct {
		anon struct {
			Id int `db:"id"` //will be ignore
		}

		ID   int    `db:"id"`
		Name string `db:"name"`
		Age  int    `db:"age"`
		RowInfo
	}

	const query = `SELECT 10 as id, 
	'test' as name, 
	 42 as age, 
	'2012-03-04 10:11:12'::timestamp as created_at, 
	'2012-03-04 10:11:13'::timestamp as updated_at,
	NULL::timestamp as deleted_at,
    21 as version`

	var person Person
	err = Get(ctx, pool, &person, query)
	noError(t, err)

	equal(t, 10, person.ID)
	equal(t, "test", person.Name)
	equal(t, 42, person.Age)
	equal(t, true, time.Date(2012, 3, 4, 10, 11, 12, 0, time.UTC).Equal(person.CreatedAt))
	equal(t, true, time.Date(2012, 3, 4, 10, 11, 13, 0, time.UTC).Equal(*person.UpdatedAt))
}

func boolPtr(v bool) *bool {
	return &v
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func stringPtr(s string) *string {
	return &s
}

func equal(t *testing.T, l, r interface{}) {
	if !reflect.DeepEqual(l, r) {
		fmt.Printf("%v not equal %v\n", js(l), js(r))
		t.Fail()
	}
}

func js(v interface{}) string {
	result, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(result)
}

func noError(t *testing.T, err error) {
	if err != nil {
		t.Fail()
		panic(err)
	}
}

func notNilError(t *testing.T, err error) {
	if err == nil {
		t.Fail()
		panic("error is nil")
	}
}

func errorContains(t *testing.T, err error, str string) {
	if err == nil {
		t.Fail()
		panic("error is nil")
	}
	if !strings.Contains(err.Error(), str) {
		t.Fail()
		panic(fmt.Sprintf("error %v does not contains %s", err, str))
	}
}
