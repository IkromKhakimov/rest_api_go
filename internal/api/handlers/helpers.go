package handlers

import (
	"errors"
	"reflect"
	"restapi/pkg/utils"
	"strings"
)

func CheckBlankFields(teacher interface{}) error {
	val := reflect.ValueOf(teacher)
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.Kind() == reflect.String && field.String() == "" {
			return utils.ErrorHandler(errors.New("All fields are required"), "All fields are required")
		}
	}
	return nil
}

func GetFieldNames(model interface{}) []string {
	val := reflect.TypeOf(model)
	fields := []string{}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldToAdd := strings.TrimSuffix(field.Tag.Get("json"), ",omitempty")
		fields = append(fields, fieldToAdd)
	}
	return fields
}
