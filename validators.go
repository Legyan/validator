package validator

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
)

var ErrNotStruct = errors.New("wrong argument given, should be a struct")
var ErrInvalidValidatorSyntax = errors.New("invalid validator syntax")
var ErrValidateForUnexportedFields = errors.New("validation for unexported field is not allowed")

type ValidationError struct {
	Err   error
	Field string
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var errorStrings []string
	for _, ve := range v {
		var field string
		if !(errors.Is(ve.Err, ErrInvalidValidatorSyntax) || errors.Is(ve.Err, ErrValidateForUnexportedFields)) {
			field = ve.Field + ": "
		}
		errorStrings = append(errorStrings, field+ve.Err.Error())
	}
	return strings.Join(errorStrings, "; ")
}

func Validate(v any) error {
	var validationErrs ValidationErrors
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		vt := field.Tag.Get("validate")
		if vt == "" {
			continue
		}
		if !field.IsExported() {
			validationErrs = append(validationErrs, ValidationError{
				Err:   ErrValidateForUnexportedFields,
				Field: field.Name,
			})
			continue
		}

		fieldValue := val.Field(i)

		rule := strings.SplitN(vt, ":", 2)
		if len(rule) != 2 {
			validationErrs = append(validationErrs, ValidationError{
				Err:   ErrInvalidValidatorSyntax,
				Field: field.Name,
			})
			continue
		}

		validatorName := rule[0]
		validatorParams := rule[1]

		validator, ok := validators[validatorName]
		if !ok {
			validationErrs = append(validationErrs, ValidationError{
				Err:   ErrInvalidValidatorSyntax,
				Field: field.Name,
			})
			continue
		}

		switch field.Type.Kind() {
		case reflect.Slice:
			elemType := field.Type.Elem().Kind()
			if elemType == reflect.Int || elemType == reflect.String {
				slice := fieldValue
				for i := 0; i < slice.Len(); i++ {
					elem := slice.Index(i)
					if err := validator(elem, validatorParams); err != nil {
						validationErrs = append(validationErrs, ValidationError{
							Err:   err,
							Field: fmt.Sprintf("%s[%d]", field.Name, i),
						})
					}
				}
			} else {
				validationErrs = append(validationErrs, ValidationError{
					Err: fmt.Errorf("unsupported element type in slice"),
				})
			}
		default:
			if err := validator(fieldValue, validatorParams); err != nil {
				validationErrs = append(validationErrs, ValidationError{
					Err:   err,
					Field: field.Name,
				})
			}
		}
	}

	if len(validationErrs) > 0 {
		return validationErrs
	}
	return nil
}

type validatorFunc func(fieldValue reflect.Value, params string) error

var validators = map[string]validatorFunc{
	"len": validateLen,
	"in":  validateIn,
	"min": validateMin,
	"max": validateMax,
}

func validateLen(fieldValue reflect.Value, params string) error {
	expectedLen, err := strconv.Atoi(params)
	if err != nil {
		return ErrInvalidValidatorSyntax
	}

	switch fieldValue.Kind() {
	case reflect.String:
		if fieldValue.Len() != expectedLen {
			return fmt.Errorf("length must be %d", expectedLen)
		}
	default:
		return fmt.Errorf("unsupported type for len validator")
	}

	return nil
}

func validateIn(fieldValue reflect.Value, params string) error {
	if len(params) == 0 {
		return ErrInvalidValidatorSyntax
	}
	values := strings.Split(params, ",")

	switch fieldValue.Kind() {
	case reflect.String:
		v := fieldValue.String()
		for _, val := range values {
			if v == val {
				return nil
			}
		}
	case reflect.Int:
		v := fieldValue.Int()
		for _, val := range values {
			intVal, err := strconv.Atoi(val)
			if err != nil {
				return ErrInvalidValidatorSyntax
			}
			if v == int64(intVal) {
				return nil
			}
		}
	default:
		return fmt.Errorf("unsupported type for in validator")
	}

	return fmt.Errorf("value not in the allowed set")
}

func validateMin(v reflect.Value, param string) error {
	min, err := strconv.Atoi(param)
	if err != nil {
		return ErrInvalidValidatorSyntax
	}

	switch v.Kind() {
	case reflect.String:
		if v.Len() < min {
			return fmt.Errorf("length must be at least %d", min)
		}
	case reflect.Int:
		if v.Int() < int64(min) {
			return fmt.Errorf("value must be at least %d", min)
		}
	default:
		return fmt.Errorf("unsupported type for min validator")
	}

	return nil
}

func validateMax(v reflect.Value, param string) error {
	max, err := strconv.Atoi(param)
	if err != nil {
		return ErrInvalidValidatorSyntax
	}

	switch v.Kind() {
	case reflect.String:
		if v.Len() > max {
			return fmt.Errorf("length must not exceed %d", max)
		}
		if v.Len() < 1 {
			return fmt.Errorf("value must be at least 1")
		}
	case reflect.Int:
		if v.Int() > int64(max) {
			return fmt.Errorf("value must not exceed %d", max)
		}
	default:
		return fmt.Errorf("unsupported type for max validator")
	}

	return nil
}
