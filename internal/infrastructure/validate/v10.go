package validate

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

// PlaygroundV10 Validator implementation using go-playground
type PlaygroundV10 struct {
	core  *validator.Validate
	trans ut.Translator
}

// NewValidator create a new Validator
func NewValidator() *PlaygroundV10 {
	en := en.New()
	zh := zh.New()
	uni := ut.New(en, en, zh)
	trans, _ := uni.GetTranslator("en") // en translator as default

	validate := validator.New()
	en_translations.RegisterDefaultTranslations(validate, trans)
	zh_translations.RegisterDefaultTranslations(validate, trans)
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
	return &PlaygroundV10{
		core:  validate,
		trans: trans,
	}
}

// Struct validate struct
func (v PlaygroundV10) Struct(s interface{}) []*FieldError {
	var result []*FieldError
	validate := v.core
	if err := validate.Struct(s); err != nil {
		for _, item := range err.(validator.ValidationErrors) {
			result = append(result, NewFieldError(item.Field(), item.Translate(v.trans)))
		}
		return result
	}
	return nil
}

// Empty check if value is empty
func (v PlaygroundV10) Empty(varName string, s interface{}) []*FieldError {
	validate := v.core
	var result []*FieldError
	if err := validate.Var(s, "required"); err != nil {
		for range err.(validator.ValidationErrors) {
			msg := fmt.Sprintf("%s is required", varName)
			result = append(result, NewFieldError(varName, msg))
		}
		return result
	}
	return nil
}

// AllEmpty check if all fields are empty
//
// names and fields have one to one relationship respect to the order
func (v PlaygroundV10) AllEmpty(names []string, fields ...interface{}) *FieldError {
	if len(names) != len(fields) {
		panic(fmt.Errorf("number of name: %d, fields: %d", len(names), len(fields)))
	}

	validate := v.core
	nonEmpty := false
	for _, s := range fields {
		if err := validate.Var(s, "required"); err == nil {
			nonEmpty = true
			break
		}
	}
	if !nonEmpty {
		return NewFieldError(strings.Join(names, ","), "One of the fields should not be empty")
	}
	return nil
}
