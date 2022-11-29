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

// Do performs a request using resty.Client with
// request of type `Request[T]`,
// response of type `T`,
// error of type `error`.
//
// Requests without response should specify the `any` type.
// Requests without typed error should specify the `error` type.
func Do[R Request[T], T any, E error](client *resty.Client, req R) (t T, err error) {
	var (
		method      = req.Method()
		path        = req.Path()
		pathParams  = req.PathParams()
		queryParams = req.QueryParams()
		headers     = req.Headers()
	)
	body, err := req.Body()
	if err != nil {
		return t, err
	}

	builder := client.R().
		SetHeaders(headers).
		SetQueryParams(queryParams).
		SetPathParams(pathParams)

	// SetBody does not handle nil
	if body != nil {
		builder = builder.SetBody(body)
	}

	res, err := builder.
		Execute(method, path)
	if err != nil {
		return t, err
	}

	var status = res.StatusCode()

	if status >= 200 && status < 300 {
		return req.ResponseFromBytes(res.Body())
	}

	if status >= 400 {
		if strings.HasPrefix(strings.ToLower(res.Header().Get("content-type")), "application/json") {
			var e E
			json.Unmarshal(res.Body(), &e)
			if err != nil {
				return t, err
			}
			return t, e
		}

		// Return body as error string
		err = fmt.Errorf("%s", res.Body())
		return t, err
	}

	return t, nil
}
