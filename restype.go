package restype

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
)

// Request[T] represents a request with a response of type T.
type Request[T any] interface {
	Method() string
	Path() string
	PathParams() map[string]string
	QueryParams() map[string]string
	Headers() map[string]string
	Body() ([]byte, error)
	ResponseFromBytes([]byte) (T, error)
}

// RequestOptions are functions that modify resty.Request and return a reference to it.
//
// Allows for modifying the request where required or
// where it does not make sense to add to implementations of Request[T].
type RequestOptions func(*resty.Request) *resty.Request

// Do performs a request using resty.Client with
// request of type `Request[T]`,
// response of type `T`,
// error of type `error`.
//
// Requests without response should specify the `any` type.
// Requests without typed error should specify the `error` type.
//
// RequestOptions enable modification of the request.
func Do[R Request[T], T any, E error](client *resty.Client, req R, opts ...RequestOptions) (t T, err error) {
	_, t, err = DoRaw[R, T, E](client, req, opts...)
	return t, err
}

// DoRaw performs the same function as Do
// and returns the raw resty.Response.
func DoRaw[R Request[T], T any, E error](client *resty.Client, req R, opts ...RequestOptions) (res *resty.Response, t T, err error) {
	var (
		method      = req.Method()
		path        = req.Path()
		pathParams  = req.PathParams()
		queryParams = req.QueryParams()
		headers     = req.Headers()
	)
	body, err := req.Body()
	if err != nil {
		return res, t, Error{kind: ErrorBodyFromRequest, original: err}
	}

	builder := client.R().
		SetHeaders(headers).
		SetQueryParams(queryParams).
		SetPathParams(pathParams)

	for _, opt := range opts {
		builder = opt(builder)
	}

	// SetBody does not handle nil
	if body != nil {
		builder = builder.SetBody(body)
	}

	res, err = builder.
		Execute(method, path)
	if err != nil {
		return res, t, Error{kind: ErrorExecuteRequest, original: err}
	}

	if res.IsSuccess() {
		t, err = req.ResponseFromBytes(res.Body())
		if err != nil {
			return res, t, Error{kind: ErrorResponseFromBytes, original: err}
		}
		return res, t, nil
	}

	if res.IsError() {
		if strings.HasPrefix(strings.ToLower(res.Header().Get("content-type")), "application/json") {
			var e E
			if err := json.Unmarshal(res.Body(), &e); err == nil {
				return res, t, e
			}
		}

		// Return body as error string
		err = fmt.Errorf("%s", res.Body())
		return res, t, err
	}

	return res, t, nil
}
