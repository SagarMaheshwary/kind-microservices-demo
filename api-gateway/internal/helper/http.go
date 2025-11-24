package helper

import (
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var (
	MessageBadRequest = "Bad Request"
)

func PrepareResponse(message string, data any) gin.H {
	return gin.H{
		"message": message,
		"data":    data,
	}
}

func PrepareResponseFromValidationError(err error, obj any) gin.H {
	errorsMap := map[string][]string{}

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, e := range ve {
			f, _ := reflect.TypeOf(obj).Elem().FieldByName(e.Field())
			field, _ := f.Tag.Lookup("json")
			errorsMap[field] = []string{ValidationErrorByTag(e.Tag(), field)}
		}

		// Add empty slices for fields without errors to keep structure consistent
		fields := reflect.VisibleFields(reflect.Indirect(reflect.ValueOf(obj)).Type())
		for _, field := range fields {
			t, _ := field.Tag.Lookup("json")
			if _, ok := errorsMap[t]; !ok {
				errorsMap[t] = []string{}
			}
		}

		return PrepareResponse(MessageBadRequest, gin.H{
			"errors": errorsMap,
		})
	}

	return PrepareResponse(MessageBadRequest, gin.H{
		"errors": errorsMap,
	})
}

func ValidationErrorByTag(tag string, field string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "invalid":
		return fmt.Sprintf("%s is invalid", field)
	case "email":
		return fmt.Sprintf("%s must be an email", field)
	}
	return ""
}
