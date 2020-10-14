package infra

// FieldError field error to be nested by other errors
type FieldError struct {
	Domain string `json:"domain"`
	Reason string `json:"reason"`
}

// RESTStandardError response error
type RESTStandardError struct {
	Type    string `json:"type,omitempty"`
	Code    int    `json:"code"`
	Title   string `json:"title"`
	Detail  string `json:"detail,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
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

func (re RESTStandardError) SetTraceID(traceID string) RESTStandardError {
	re.TraceID = traceID
	return re
}

// RESTValidationError standard validation error
type RESTValidationError struct {
	RESTStandardError
	InvalidParams []*FieldError `json:"invalid_params"`
}

func NewRESTValidationError(code int, title string, internal []*FieldError) *RESTValidationError {
	return &RESTValidationError{
		RESTStandardError: RESTStandardError{
			Code:  code,
			Title: title,
		},
		InvalidParams: internal,
	}
}

func (rve RESTValidationError) Error() string {
	return rve.Detail
}

func (rve RESTValidationError) SetDetail(detail string) RESTValidationError {
	rve.Detail = detail
	return rve
}

func (rve RESTValidationError) SetTraceID(traceID string) RESTValidationError {
	rve.RESTStandardError.TraceID = traceID
	return rve
}
