package easyscan

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func Test_getStructDBTags(t *testing.T) {
	t.Parallel()
	type testType struct {
		Field0 int `db:"Field0"`
		Field1 int `db:"Field1"`
		Field2 int `db:"Field2"`
		Field3 int `db:"Field3"`
		Field4 int `db:"Field4"`
	}

	type bigTestType struct {
		Field0  int `db:"Field0"`
		Field1  int `db:"Field1"`
		Field2  int `db:"Field2"`
		Field3  int `db:"Field3"`
		Field4  int `db:"Field4"`
		Field5  int `db:"Field5"`
		Field6  int `db:"Field6"`
		Field7  int `db:"Field7"`
		Field8  int `db:"Field8"`
		Field9  int `db:"Field9"`
		Field10 int `db:"Field10"`
		Field11 int `db:"Field11"`
		Field12 int `db:"Field12"`
		Field13 int `db:"Field13"`
		Field14 int `db:"Field14"`
		Field15 int `db:"Field15"`
		Field16 int `db:"Field16"`
		Field17 int `db:"Field17"`
		Field18 int `db:"Field18"`
		Field19 int `db:"Field19"`
		Field20 int `db:"Field20"`
	}

	t.Run("concurrency", func(t *testing.T) {
		const routines = 1
		const iterations = 20

		wg := &sync.WaitGroup{}
		wg.Add(routines)
		for th := 0; th < routines; th++ {
			go func() {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					tags := getStructDBTags(reflect.TypeOf(testType{}))
					for f := 0; f < 5; f++ {
						idx, ok := tags.find([]byte(fmt.Sprintf("Field%d", f)))
						equal(t, true, ok)
						equal(t, f, idx)
					}

					bigTypeTags := getStructDBTags(reflect.TypeOf(bigTestType{}))
					for f := 0; f < 21; f++ {
						idx, ok := bigTypeTags.find([]byte(fmt.Sprintf("Field%d", f)))
						equal(t, true, ok)
						equal(t, f, idx)
					}
				}
			}()
		}
		wg.Wait()
	})

	t.Run("not found", func(t *testing.T) {
		tags := getStructDBTags(reflect.TypeOf(testType{}))
		idx, ok := tags.find([]byte("test_123"))
		equal(t, false, ok)
		equal(t, 0, idx)

		tags = getStructDBTags(reflect.TypeOf(bigTestType{}))
		idx, ok = tags.find([]byte("test_123"))
		equal(t, false, ok)
		equal(t, 0, idx)
	})

}
