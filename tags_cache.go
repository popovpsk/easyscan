package easyscan

import (
	"reflect"
	"sync"
)

const (
	dbTagName     = "db"
	tagsThreshold = 16
)

var tagsCache = new(sync.Map)

type tagsContainer interface {
	find(column []byte) (fieldIdx int, ok bool)
}

type sliceContainer []structField

func createSliceContainer(fields []structField) sliceContainer {
	return fields
}

func (s sliceContainer) find(column []byte) (int, bool) {
	for _, v := range s {
		if v.tag == string(column) {
			return v.idx, true
		}
	}
	return 0, false
}

type mapContainer map[string]int

func createMapContainer(fields []structField) mapContainer {
	m := make(map[string]int, len(fields))
	for _, v := range fields {
		m[v.tag] = v.idx
	}
	return m
}

func (s mapContainer) find(column []byte) (int, bool) {
	idx, ok := s[string(column)]
	return idx, ok
}

type structField struct {
	idx int
	tag string
}

func extractDbTags(t reflect.Type) []structField {
	numField := t.NumField()
	fields := make([]string, numField)
	cnt := 0
	for i := 0; i < numField; i++ {
		tag := t.Field(i).Tag.Get(dbTagName)
		if tag != "" {
			cnt++
		}
		fields[i] = tag
	}

	result := make([]structField, 0, cnt)
	for i, v := range fields {
		if v != "" {
			result = append(result, structField{
				idx: i,
				tag: v,
			})
		}
	}

	return result
}

func getStructDBTags(t reflect.Type) tagsContainer {
	tags, ok := tagsCache.Load(t)
	if ok {
		return tags.(tagsContainer)
	}

	fields := extractDbTags(t)
	var result tagsContainer
	if len(fields) > tagsThreshold {
		result = createMapContainer(fields)
	} else {
		result = createSliceContainer(fields)
	}

	tagsCache.Store(t, result)
	return result
}
