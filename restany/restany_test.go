package restany

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Jiraffe7/restype"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

// AccountRequest implementing restype.Request is the typical way to make a typed request/response pair
// for use with restype.

var _ restype.Request[AccountResponse] = new(AccountRequest)

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

// AccountRequestWrapped is the wrapped type for types that can be used with restype.Do.
// A wrapped type is required to use restany.Do.
type AccountRequestWrapped struct {
	*AccountRequest
}

// ResponseFromBytes that returns (interface{}, error)
// needs to be implemented for wrapped types to be used with restany.Do.
func (s *AccountRequestWrapped) ResponseFromBytes(bs []byte) (interface{}, error) {
	return s.AccountRequest.ResponseFromBytes(bs)
}

func TestDo_OK(t *testing.T) {
	var req = AccountRequest{
		Name:   "account1234",
		Token:  "token1234",
		ID:     42,
		Status: "active",
	}
	var accountResponse = AccountResponse{Users: []User{{1, "A"}, {2, "B"}}}

	var handled = false
	var srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	}))
	defer srv.Close()

	var client = resty.New().
		SetBaseURL(srv.URL)

	var wrappedReq = AccountRequestWrapped{&req}
	res, err := Do(client, &wrappedReq)
	assert.NoError(t, err)
	assert.True(t, handled)

	assert.Equal(t, accountResponse, res)
}
