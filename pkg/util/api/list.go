package api

import (
	"errors"
	"fmt"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjAsPtr[T any] interface {
	client.Object
	*T
}

type ObjListAsPtr[T any] interface {
	client.ObjectList
	*T
}

func GetItemsFromList[E any](l client.ObjectList) ([]E, error) {
	listValue := reflect.ValueOf(l)
	for k := listValue.Kind(); k == reflect.Interface || k == reflect.Pointer; k = listValue.Kind() {
		listValue = listValue.Elem()
	}
	if listValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("list kind %v is not a struct", listValue.Kind())
	}

	itemsValue := listValue.FieldByName("Items")
	if (itemsValue == reflect.Value{}) {
		return nil, errors.New("no Items member found")
	}

	if itemsValue.Kind() != reflect.Slice {
		return nil, errors.New(`"Items" member is not a slice`)
	}

	ret, canConvert := itemsValue.Interface().([]E)
	if !canConvert {
		return nil, fmt.Errorf("%v items cannot be converted to %v", itemsValue.Type(), reflect.TypeFor[[]E]())
	}

	return ret, nil
}
