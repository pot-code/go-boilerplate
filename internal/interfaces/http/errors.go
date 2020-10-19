package http

import (
	"net/http"

	infra "github.com/pot-code/go-boilerplate/internal/infrastructure"
)

// RESTStandardError response error
type RESTStandardError struct {
	Type    string `json:"type,omitempty"`
	Code    int    `json:"code"`
	Title   string `json:"title"`
	Detail  string `json:"detail,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
}

func NewRESTStandardError(code int, detail string) *RESTStandardError {
	return &RESTStandardError{
		Code:   code,
		Title:  http.StatusText(code),
		Detail: detail,
	}
}

func (re RESTStandardError) Error() string {
	return re.Detail
}

func (re RESTStandardError) SetTraceID(traceID string) RESTStandardError {
	re.TraceID = traceID
	return re
}

// RESTValidationError standard validation error
type RESTValidationError struct {
	RESTStandardError
	InvalidParams []*infra.FieldError `json:"invalid_params"`
}

func NewRESTValidationError(code int, detail string, internal []*infra.FieldError) *RESTValidationError {
	return &RESTValidationError{
		RESTStandardError: RESTStandardError{
			Code:   code,
			Title:  http.StatusText(code),
			Detail: detail,
		},
		InvalidParams: internal,
	}
}

func (rve RESTValidationError) Error() string {
	return rve.Detail
}

func (rve RESTValidationError) SetTraceID(traceID string) RESTValidationError {
	rve.RESTStandardError.TraceID = traceID
	return rve
}
