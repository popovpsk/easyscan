package easyscan

import (
	"reflect"
	"sync"
)

const (
	dbTagName           = "db"
	sliceContainerLimit = 32
)

type fieldsContainer interface {
	find(column []byte, value reflect.Value) interface{}
}

type fieldsContainerSlice []structField
type fieldsContainerMap map[string]fieldPath

// intrusive linked list
type fieldPath struct {
	idx  int
	next *fieldPath
}

type structField struct {
	idx   fieldPath
	dbTag string
}

func createSliceContainer(fields []structField) fieldsContainerSlice {
	return fields
}

func (s fieldsContainerSlice) find(column []byte, value reflect.Value) interface{} {
	for _, v := range s {
		if v.dbTag == string(column) {
			return v.idx.eface(value)
		}
	}
	return emptyScanObj
}

func createMapContainer(fields []structField) fieldsContainerMap {
	m := make(map[string]fieldPath, len(fields))
	for _, v := range fields {
		m[v.dbTag] = v.idx
	}
	return m
}

func (s fieldsContainerMap) find(column []byte, value reflect.Value) interface{} {
	idx, ok := s[string(column)]
	if !ok {
		return emptyScanObj
	}
	return idx.eface(value)
}

func (f *fieldPath) eface(t reflect.Value) interface{} {
	next := f

	for {
		field := t.Field((*next).idx)
		if next.next == nil {
			return field.Addr().Interface()
		} else {
			t = field
			next = next.next
		}
	}
}

func (f *fieldPath) append(e *fieldPath) {
	for f.next != nil {
		f = f.next
	}
	f.next = e
}

func (f *fieldPath) copy() fieldPath {
	result := *f

	l := *f
	r := &result

	for l.next != nil {
		next := *l.next

		r.next = &next
		r = r.next
		l = next
	}
	return result
}

func extractFields(t reflect.Type) []structField {
	fields := make([]structField, 0)
	exploreStruct(t, &fields, nil)

	return fields[:len(fields):len(fields)]
}

func exploreStruct(t reflect.Type, result *[]structField, root *fieldPath) {
	numField := t.NumField()

	for i := 0; i < numField; i++ {
		f := t.Field(i)
		tag := f.Tag.Get(dbTagName)

		if tag != "" {
			sf := structField{dbTag: tag}
			fieldIndex := fieldPath{idx: i}

			if root == nil {
				sf.idx = fieldIndex
			} else {
				cp := root.copy()
				cp.append(&fieldIndex)
				sf.idx = cp
			}

			*result = append(*result, sf)
			continue
		}

		//embedding here
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			if root == nil {
				r := fieldPath{idx: i}
				exploreStruct(f.Type, result, &r)
			} else {
				cp := root.copy()
				cp.append(&fieldPath{idx: i})
				exploreStruct(f.Type, result, &cp)
			}
		}
	}
}

var typeCache = new(sync.Map)

func getTaggedFields(t reflect.Type) fieldsContainer {
	cached, ok := typeCache.Load(t)
	if ok {
		return cached.(fieldsContainer)
	}

	fields := extractFields(t)
	var result fieldsContainer
	if len(fields) > sliceContainerLimit {
		result = createMapContainer(fields)
	} else {
		result = createSliceContainer(fields)
	}

	typeCache.Store(t, result)
	return result
}
