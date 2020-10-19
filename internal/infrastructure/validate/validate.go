package validate

// FieldError field error to be nested by other errors
type FieldError struct {
	Domain string `json:"domain"`
	Reason string `json:"reason"`
}

// NewFieldError create new field error
func NewFieldError(domain string, reason string) *FieldError {
	return &FieldError{domain, reason}
}

// Validator .
type Validator interface {
	Struct(s interface{}) []*FieldError
	Empty(varName string, s interface{}) []*FieldError
	AllEmpty(names []string, fields ...interface{}) *FieldError
}
