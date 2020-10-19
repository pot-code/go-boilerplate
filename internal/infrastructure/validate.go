package infra

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

// FieldError field error to be nested by other errors
type FieldError struct {
	Domain string `json:"domain"`
	Reason string `json:"reason"`
}

func NewFieldError(domain string, reason string) *FieldError {
	return &FieldError{domain, reason}
}

// Validator .
type Validator interface {
	Struct(s interface{}) []*FieldError
	Empty(varName string, s interface{}) []*FieldError
	AllEmpty(names []string, fields ...interface{}) *FieldError
}

// ValidatorV10 Validator implementation using go-playground
type ValidatorV10 struct {
	core  *validator.Validate
	trans ut.Translator
}

// NewValidator create a new Validator
func NewValidator() *ValidatorV10 {
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
	return &ValidatorV10{
		core:  validate,
		trans: trans,
	}
}

// Struct validate struct
func (v ValidatorV10) Struct(s interface{}) []*FieldError {
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
func (v ValidatorV10) Empty(varName string, s interface{}) []*FieldError {
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
func (v ValidatorV10) AllEmpty(names []string, fields ...interface{}) *FieldError {
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
