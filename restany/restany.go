package restany

import (
	"github.com/Jiraffe7/restype"
	"github.com/go-resty/resty/v2"
)

// Do is a non-generic version of restype.Do.
// Supports the use of restype with go versions before 1.18.
func Do(client *resty.Client, req restype.Request[interface{}]) (t interface{}, err error) {
	_, t, err = restype.DoRaw[restype.Request[interface{}], any, error](client, req)
	return t, err
}
