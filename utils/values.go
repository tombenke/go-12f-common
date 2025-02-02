package utils

import (
	"net/url"
	"reflect"
)

func IsEmpty(value interface{}) bool {
	if value == nil {
		return true
	}
	switch realValue := value.(type) {
	case string:
		return realValue == ""
	case []interface{}:
		return len(realValue) == 0
	default:
		reflectValue := reflect.ValueOf(value)
		if reflectValue.Kind() == reflect.Ptr && reflectValue.IsNil() {
			return true
		}
		return reflectValue.Kind() == reflect.Slice && reflectValue.Len() == 0
	}
}

func SafeUrl(u *url.URL) *url.URL {
	if u == nil {
		u = &url.URL{}
	}
	return u
}
