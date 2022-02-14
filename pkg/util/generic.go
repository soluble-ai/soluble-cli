package util

import (
	"reflect"
	"strings"
)

func GenericSet(mp *map[string]interface{}, path string, value interface{}) {
	if *mp == nil {
		m := make(map[string]interface{})
		*mp = m
	}
	m := *mp
	pathElements := strings.Split(path, "/")
	for len(pathElements) > 1 {
		name := pathElements[0]
		val := m[name]
		if val == nil || reflect.ValueOf(val).IsNil() {
			sm := make(map[string]interface{})
			m[name] = sm
			m = sm
		} else {
			m = val.(map[string]interface{})
		}
		pathElements = pathElements[1:]
	}
	m[pathElements[0]] = value
}

func GenericGet(m map[string]interface{}, path string) interface{} {
	pathElements := strings.Split(path, "/")
	for len(pathElements) > 1 {
		name := pathElements[0]
		m = m[name].(map[string]interface{})
		pathElements = pathElements[1:]
	}
	return m[pathElements[0]]
}
