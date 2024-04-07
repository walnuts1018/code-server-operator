package common

import (
	"fmt"
	"reflect"
)

func Parse[T any](obj *T, params map[string]string) error {
	t := reflect.TypeOf(*obj)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if field.Type.Kind() != reflect.String {
			continue
		}
		json, exist := field.Tag.Lookup("json")
		if !exist {
			json = field.Name
		}

		value, ok := params[json]
		if ok {
			reflect.ValueOf(obj).Elem().Field(i).SetString(value)
		} else {
			text, exist := field.Tag.Lookup("required")
			if exist && text == "true" {
				return fmt.Errorf("field %s is required", field.Name)
			}
		}
	}

	return nil
}
