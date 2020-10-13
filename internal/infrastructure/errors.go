package infra

// FieldError field error to be nested by other errors
type FieldError struct {
	Domain string `json:"domain"`
	Reason string `json:"reason"`
}

// RESTStandardError response error
type RESTStandardError struct {
	Type   string `json:"type,omitempty"`
	Code   int    `json:"code"`
	Title  string `json:"title"`
	Detail string `json:"detail,omitempty"`
}

func NewRESTStandardError(code int, title string) *RESTStandardError {
	return &RESTStandardError{
		Code:  code,
		Title: title,
	}
}

func (re RESTStandardError) Error() string {
	return re.Detail
}

func (re RESTStandardError) SetDetail(detail string) RESTStandardError {
	re.Detail = detail
	return re
}

// RESTValidationError standard validation error
type RESTValidationError struct {
	RESTStandardError
	Errors []*FieldError `json:"errors"`
}

func NewRESTValidationError(code int, title string, internal []*FieldError) *RESTValidationError {
	return &RESTValidationError{
		RESTStandardError: RESTStandardError{
			Code:  code,
			Title: title,
		},
		Errors: internal,
	}
}

func (rve RESTValidationError) Error() string {
	return rve.Detail
}

func (rve RESTValidationError) SetDetail(detail string) RESTValidationError {
	rve.Detail = detail
	return rve
}
