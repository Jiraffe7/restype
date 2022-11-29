package restype

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

var _ Request[AccountResponse] = new(AccountRequest)

type AccountRequest struct {
	// Body
	Name string `json:"name"`

	// Headers
	Token string `json:"-"`

	// Path params
	ID int `json:"-"`

	// Query params
	Status string `json:"-"`
}

// Body implements Request
func (s *AccountRequest) Body() ([]byte, error) {
	return json.Marshal(s)
}

// Headers implements Request
func (s *AccountRequest) Headers() map[string]string {
	var headers = make(map[string]string)
	headers["x-account-token"] = s.Token
	return headers
}

// Method implements Request
func (*AccountRequest) Method() string {
	return "POST"
}

// Path implements Request
func (*AccountRequest) Path() string {
	return "/api/account/{account_id}"
}

// PathParams implements Request
func (s *AccountRequest) PathParams() map[string]string {
	var params = make(map[string]string)
	params["account_id"] = fmt.Sprint(s.ID)
	return params
}

// QueryParams implements Request
func (s *AccountRequest) QueryParams() map[string]string {
	var params = make(map[string]string)
	params["account_status"] = s.Status
	return params
}

// ResponseFromBytes implements Request
func (*AccountRequest) ResponseFromBytes(bs []byte) (AccountResponse, error) {
	var res AccountResponse
	err := json.Unmarshal(bs, &res)
	if err != nil {
		return AccountResponse{}, err
	}
	return res, nil
}

type AccountResponse struct {
	Users []User
}

type User struct {
	ID   int
	Name string
}

var _ error = new(CustomError)

type CustomError struct {
	ID      int
	Message string
	Args    map[string]any
}

// Error implements error
func (s CustomError) Error() string {
	return fmt.Sprintf("CustomError(ID=%d): %s %v", s.ID, s.Message, s.Args)
}

func TestDo_OK(t *testing.T) {
	var req = AccountRequest{
		Name:   "account1234",
		Token:  "token1234",
		ID:     42,
		Status: "active",
	}
	var accountResponse = AccountResponse{Users: []User{{1, "A"}, {2, "B"}}}

	var mux = http.NewServeMux()
	var handled = false
	mux.HandleFunc("/api/account/", func(w http.ResponseWriter, r *http.Request) {
		handled = true
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/account/42", r.URL.Path)
		assert.Equal(t, "active", r.URL.Query().Get("account_status"))
		assert.Equal(t, "token1234", r.Header.Get("x-account-token"))

		var reqBody AccountRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NoError(t, err)
		assert.Equal(t, AccountRequest{Name: "account1234"}, reqBody)

		err = json.NewEncoder(w).Encode(&accountResponse)
		assert.NoError(t, err)
	})
	var srv = http.Server{Addr: ":8080", Handler: mux}
	go srv.ListenAndServe()
	defer srv.Shutdown(context.Background())

	var client = resty.New().
		SetBaseURL("http://localhost:8080")

	res, err := Do[*AccountRequest, AccountResponse, CustomError](client, &req)
	assert.NoError(t, err)
	assert.True(t, handled)

	assert.Equal(t, accountResponse, res)
}

func TestDo_CustomErrorResponse(t *testing.T) {
	var req = AccountRequest{
		Name:   "account1234",
		Token:  "token1234",
		ID:     42,
		Status: "active",
	}
	var customError = CustomError{
		ID:      22,
		Message: "error with request",
		Args: map[string]any{
			"arg1": float64(33),
			"arg2": "asdf",
		},
	}

	var mux = http.NewServeMux()
	var handled = false
	mux.HandleFunc("/api/account/", func(w http.ResponseWriter, _ *http.Request) {
		handled = true

		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(&customError)
		assert.NoError(t, err)
	})
	var srv = http.Server{Addr: ":8080", Handler: mux}
	go srv.ListenAndServe()
	defer srv.Shutdown(context.Background())

	var client = resty.New().
		SetBaseURL("http://localhost:8080")

	_, err := Do[*AccountRequest, AccountResponse, CustomError](client, &req)
	assert.True(t, handled)
	assert.Equal(t, customError, err)
}

func TestDo_CustomErrorResponseNotJSON(t *testing.T) {
	var req = AccountRequest{
		Name:   "account1234",
		Token:  "token1234",
		ID:     42,
		Status: "active",
	}
	var customError = CustomError{
		ID:      22,
		Message: "error with request",
		Args: map[string]any{
			"arg1": float64(33),
			"arg2": "asdf",
		},
	}

	var mux = http.NewServeMux()
	var handled = false
	mux.HandleFunc("/api/account/", func(w http.ResponseWriter, _ *http.Request) {
		handled = true

		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(&customError)
		assert.NoError(t, err)
	})
	var srv = http.Server{Addr: ":8080", Handler: mux}
	go srv.ListenAndServe()
	defer srv.Shutdown(context.Background())

	var client = resty.New().
		SetBaseURL("http://localhost:8080")

	_, resErr := Do[*AccountRequest, AccountResponse, error](client, &req)
	assert.True(t, handled)

	want, err := json.Marshal(&customError)
	assert.NoError(t, err)
	assert.Equal(t, string(want)+"\n", resErr.Error())
}

var _ Request[any] = new(RequestWithNilResponse)

type RequestWithNilResponse struct {
	Default
}

// Method implements Request
func (*RequestWithNilResponse) Method() string {
	return "POST"
}

// Path implements Request
func (*RequestWithNilResponse) Path() string {
	return "/"
}

func TestDo_NilResponseType(t *testing.T) {
	var req = RequestWithNilResponse{}

	var mux = http.NewServeMux()
	var handled = false
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		handled = true
		w.WriteHeader(http.StatusOK)
	})
	var srv = http.Server{Addr: ":8080", Handler: mux}
	go srv.ListenAndServe()
	defer srv.Shutdown(context.Background())

	var client = resty.New().
		SetBaseURL("http://localhost:8080")

	res, err := Do[*RequestWithNilResponse, any, error](client, &req)
	assert.NoError(t, err)
	assert.True(t, handled)

	assert.Nil(t, res)
}

func TestDo_NoRoute(t *testing.T) {
	var req = RequestWithNilResponse{}

	var mux = http.NewServeMux()
	var srv = http.Server{Addr: ":8080", Handler: mux}
	go srv.ListenAndServe()
	defer srv.Shutdown(context.Background())

	var client = resty.New().
		SetBaseURL("http://localhost:8080")

	res, err := Do[*RequestWithNilResponse, any, error](client, &req)
	assert.Equal(t, errors.New("404 page not found\n"), err)
	assert.Nil(t, res)

}
