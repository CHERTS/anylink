package sessdata

import (
	"fmt"
	"reflect"
)

// Overwrite a's with all fields of b
// If fields is not empty, it means that the specific fields of b are used to overwrite a's
// a should be a structure pointer
func CopyStruct(a interface{}, b interface{}, fields ...string) (err error) {
	at := reflect.TypeOf(a)
	av := reflect.ValueOf(a)
	bt := reflect.TypeOf(b)
	bv := reflect.ValueOf(b)

	// Make a simple judgment
	if at.Kind() != reflect.Ptr {
		err = fmt.Errorf("a must be a struct pointer")
		return
	}
	av = reflect.ValueOf(av.Interface())

	// which fields to copy
	_fields := make([]string, 0)
	if len(fields) > 0 {
		_fields = fields
	} else {
		for i := 0; i < bv.NumField(); i++ {
			_fields = append(_fields, bt.Field(i).Name)
		}
	}

	if len(_fields) == 0 {
		fmt.Println("no fields to copy")
		return
	}

	// copy
	for i := 0; i < len(_fields); i++ {
		name := _fields[i]
		f := av.Elem().FieldByName(name)
		bValue := bv.FieldByName(name)

		// Copy only if there are fields with the same name in a and the types are consistent.
		if f.IsValid() && f.Kind() == bValue.Kind() {
			f.Set(bValue)
		}
	}
	return
}
