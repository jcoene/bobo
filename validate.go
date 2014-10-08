// Based on the validate package at github.com/mccoyst/validate
// © 2013 Steve McCoy under the MIT license.

package bobo

import (
	"fmt"
	"reflect"
	"strings"
)

type ValidationError struct {
	Code  string `json:"code"`
	Field string `json:"field"`
	Title string `json:"title"`
}

func (e ValidationError) Error() string {
	return e.Title
}

type ValidationErrors []ValidationError

func (es ValidationErrors) Error() string {
	ss := make([]string, len(es))
	for i, e := range es {
		ss[i] = e.Error()
	}

	return strings.Join(ss, ", ")
}

type Validator map[string]func(string, interface{}) error

func (v Validator) Validate(s interface{}) error {
	val := reflect.ValueOf(s)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	t := val.Type()

	if t == nil || t.Kind() != reflect.Struct {
		return nil
	}

	ve := make(ValidationErrors, 0)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := val.Field(i)
		if !fv.CanInterface() {
			continue
		}

		val := fv.Interface()

		name := f.Tag.Get("json")
		if name == "" {
			name = f.Name
		}

		tag := f.Tag.Get("validate")
		if tag == "" {
			continue
		}

		vts := strings.Split(tag, ",")

		for _, vt := range vts {
			vf := v[vt]
			if vf == nil {
				ve = append(ve, ValidationError{fmt.Sprintf("%s_%s", name, vt), name, fmt.Sprintf("Invalid validator %s for field %s", vt, name)})
				continue
			}
			if err := vf(name, val); err != nil {
				switch err.(type) {
				case ValidationError:
					ve = append(ve, err.(ValidationError))
				default:
					ve = append(ve, ValidationError{fmt.Sprintf("%s_%s", name, vt), name, err.Error()})
				}
			}
		}
	}

	if len(ve) > 0 {
		return ve
	}

	return nil
}

func ValidatePresence(f string, i interface{}) error {
	if isEmptyValue(i) {
		return fmt.Errorf("is invalid")
	}

	return nil
}

func isEmptyValue(i interface{}) bool {
	v := reflect.ValueOf(i)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}

	return false
}
