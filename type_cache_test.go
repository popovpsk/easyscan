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
					{
						tt := reflect.TypeOf(testType{})
						tv := reflect.New(tt).Elem()
						tags := getTaggedFields(tt)
						for f := 0; f < 5; f++ {
							eface := tags.find([]byte(fmt.Sprintf("Field%d", f)), tv)
							v, ok := eface.(*int)
							equal(t, true, ok)

							*v = f
							equal(t, f, int(tv.Field(f).Int()))
						}
					}
					{
						btt := reflect.TypeOf(bigTestType{})
						btv := reflect.New(btt).Elem()
						bigType := getTaggedFields(btt)
						for f := 0; f < 21; f++ {
							eface := bigType.find([]byte(fmt.Sprintf("Field%d", f)), btv)
							v, ok := eface.(*int)
							equal(t, true, ok)

							*v = f
							equal(t, f, int(btv.Field(f).Int()))
						}
					}
				}
			}()
		}
		wg.Wait()
	})

	t.Run("not found", func(t *testing.T) {
		tt := new(testType)
		tags := getTaggedFields(reflect.TypeOf(*tt))
		f := tags.find([]byte("test_123"), reflect.ValueOf(*tt))
		equal(t, true, emptyScanObj == f.(emptyScan))

		btt := new(bigTestType)
		tags = getTaggedFields(reflect.TypeOf(*btt))
		f = tags.find([]byte("test_123"), reflect.ValueOf(*btt))
		equal(t, true, emptyScanObj == f.(emptyScan))
	})

}
