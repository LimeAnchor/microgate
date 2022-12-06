package microgate

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func (data *DataObject) GetObjectByPath(path string) *DataObject {
	pathElements := strings.Split(path, ".")
	subobject := *data
	for _, pathElement := range pathElements {
		subobject, _ = subobject.GetObject(pathElement)
	}
	return &subobject
}

func (data *DataObject) PutObjectByPath(path string) *DataObject {
	pathElements := strings.Split(path, ".")
	subobject := *data
	for _, pathElement := range pathElements {
		if subobject[pathElement] == nil {
			subobject.Put(pathElement, NewDataObject())
		}
		subobject, _ = subobject.GetObject(pathElement)
	}
	return &subobject
}

func (data *DataObject) GetString(key string) string {
	x := *data
	v := x[key]
	s := ""
	switch t := v.(type) {
	case string:
		s = t
	}
	return s
}

func (data *DataObject) GetInt(key string) int {
	x := *data
	v := x[key]
	s := 0
	switch t := v.(type) {
	case int:
		s = t
	}
	return s
}

func (data *DataObject) GetFloat64(key string) float64 {
	x := *data
	v := x[key]
	var s float64
	switch t := v.(type) {
	case float64:
		s = t
	}
	return s
}

func (data *DataObject) GetBool(key string) bool {
	x := *data
	v := x[key]
	var s bool
	switch t := v.(type) {
	case bool:
		s = t
	}
	return s
}

func MarshalToDataObject(b []byte) *DataObject {
	new := NewDataObject()
	json.Unmarshal(b, &new)
	return &new
}

type DataObject map[string]interface{}

func NewDataObject() DataObject {
	return DataObject{}
}

func (this DataObject) Put(key string, value interface{}) DataObject {
	this[key] = value
	return this
}

func (this DataObject) Get(key string) interface{} {
	return this[key]
}

func (this DataObject) GetObject(key string) (value DataObject, err error) {
	switch this[key].(type) {
	case map[string]interface{}:
		object := NewDataObject()

		for k, v := range this[key].(map[string]interface{}) {
			object.Put(k, v)
		}

		return object, nil
	case DataObject:
		return this[key].(DataObject), nil
	}

	return nil, errors.New(fmt.Sprintf("Casting error. Interface is %s, not jsongo.object", reflect.TypeOf(this[key])))
}

func (this DataObject) GetArray(key string) (newArray *List, err error) {
	newArray = Array()

	switch this[key].(type) {
	case []interface{}:
		values := this[key].([]interface{})

		for _, value := range values {
			newArray.Put(value)
		}

		return newArray, nil
	case []string:
		values := this[key].([]string)

		for _, value := range values {
			newArray.Put(value)
		}

		return newArray, nil
	case *List:
		return this[key].(*List), nil
	}

	return nil, errors.New(fmt.Sprintf("Casting error. Interface is %s, not jsongo.A or []interface{}", reflect.TypeOf(this[key])))
}

func (this DataObject) Remove(key string) DataObject {
	delete(this, key)
	return this
}

func (this DataObject) Indent() string {
	return indent(this)
}

func (this DataObject) String() string {
	return _string(this)
}

func indent(v interface{}) string {
	indent, _ := json.MarshalIndent(v, "", "   ")
	return string(indent)
}

func _string(v interface{}) string {
	indent, _ := json.Marshal(v)
	return string(indent)
}

type List []interface{}

func Array() *List {
	return &List{}
}

func (list *List) Put(value interface{}) *List {
	*list = append(*list, value)
	return list
}

func (list *List) Indent() string {
	return indent(list)
}

func (list *List) String() string {
	return _string(list)
}

func (list *List) Size() int {
	return len(*list)
}

func (list *List) OfString() (values []string, err error) {
	for _, value := range *list {
		if reflect.TypeOf(value).String() != "string" {
			return nil, errors.New(fmt.Sprintf("Value is %s, not a string.", reflect.TypeOf(value)))
		}

		values = append(values, value.(string))
	}

	return values, nil
}
