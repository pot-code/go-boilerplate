package infra

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

// Validator
type Validator interface {
	Struct(s interface{}) []*FieldError
	Empty(varName string, s interface{}) []*FieldError
	AllEmpty(names []string, fields ...interface{}) []*FieldError
}

func getValidateErrorMessage(err validator.FieldError) string {
	tag := err.Tag()
	switch tag {
	case "required":
		return "required"
	case "email":
		return "Should be in email form"
	case "min":
		return fmt.Sprintf("Length should be longer than or equal to %s", err.Param())
	case "max":
		return fmt.Sprintf("Length should be shorter than or equal to %s;", err.Param())
	}
	return tag
}

// ValidatorV10 Validator implementation using go-playground
type ValidatorV10 struct {
	core *validator.Validate
}

// NewValidator .
func NewValidator() *ValidatorV10 {
	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("json")
		if name == "-" || name == "" {
			name = fld.Tag.Get("yaml")
			if name == "-" || name == "" {
				return ""
			}
		}
		return name
	})
	return &ValidatorV10{
		core: validate,
	}
}

// Struct validate struct
func (v ValidatorV10) Struct(s interface{}) []*FieldError {
	var result []*FieldError
	validate := v.core
	if err := validate.Struct(s); err != nil {
		for _, item := range err.(validator.ValidationErrors) {
			result = append(result, &FieldError{
				Domain: item.Field(),
				Reason: getValidateErrorMessage(item),
			})
		}
		return result
	}
	return nil
}

// Empty check if value is empty
func (v ValidatorV10) Empty(varName string, s interface{}) []*FieldError {
	validate := v.core
	var result []*FieldError
	if err := validate.Var(s, "required"); err != nil {
		for range err.(validator.ValidationErrors) {
			msg := fmt.Sprintf("%s is required", varName)
			result = append(result, &FieldError{
				Domain: varName,
				Reason: msg,
			})
		}
		return result
	}
	return nil
}

// AllEmpty check if all fields are empty
//
// names and fields have one to one relationship respect to the order
func (v ValidatorV10) AllEmpty(names []string, fields ...interface{}) []*FieldError {
	if len(names) != len(fields) {
		panic(fmt.Errorf("number of name: %d, fields: %d", len(names), len(fields)))
	}

	var result []*FieldError
	validate := v.core
	nonEmpty := false
	for _, s := range fields {
		if err := validate.Var(s, "required"); err == nil {
			nonEmpty = true
			break
		}
	}
	if !nonEmpty {
		for _, name := range names {
			result = append(result, &FieldError{
				Domain: name,
				Reason: "required",
			})
		}
		return result
	}
	return nil
}
