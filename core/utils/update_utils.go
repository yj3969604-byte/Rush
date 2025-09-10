package utils

import (
	"fmt"
	"reflect"
)

func UpdateStructFromStruct(target interface{}, source interface{}) error {
	targetValue := reflect.ValueOf(target)
	sourceValue := reflect.ValueOf(source)
	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to a struct")
	}
	if sourceValue.Kind() != reflect.Struct {
		return fmt.Errorf("source must be a struct")
	}
	targetValue = targetValue.Elem()
	sourceType := sourceValue.Type()
	for i := 0; i < sourceType.NumField(); i++ {
		field := sourceType.Field(i)
		fieldName := field.Name
		targetField := targetValue.FieldByName(fieldName)
		if !targetField.IsValid() {
			continue
		}
		if !targetField.CanSet() {
			return fmt.Errorf("field %s cannot be set in target struct", fieldName)
		}
		sourceFieldValue := sourceValue.Field(i)
		switch sourceFieldValue.Kind() {
		case reflect.Bool:
			if sourceFieldValue.Bool() != targetField.Bool() {
				targetField.Set(sourceFieldValue)
			}

		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.String:
			if reflect.DeepEqual(sourceFieldValue.Interface(), reflect.Zero(sourceFieldValue.Type()).Interface()) {
				continue
			}
			if sourceFieldValue.Type().AssignableTo(targetField.Type()) {
				targetField.Set(sourceFieldValue)
			} else {
				return fmt.Errorf("type mismatch for field %s: source type %s, target type %s",
					fieldName, sourceFieldValue.Type(), targetField.Type())
			}
		default:
			// 忽略非基础类型
			continue
		}
	}
	return nil
}
