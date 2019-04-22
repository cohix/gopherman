package gopherman

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cohix/gopherman/postman"
)

// Tester represents a collection test tool
type Tester struct {
	Client     *http.Client
	Collection *postman.Collection
	Hostname   string
	Port       string
}

// NewTesterWithCollection loads a collection from a file
func NewTesterWithCollection(filepath string) (*Tester, error) {
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	collection := postman.Collection{}
	if err := json.Unmarshal(file, &collection); err != nil {
		return nil, err
	}

	tester := Tester{
		Client:     http.DefaultClient,
		Collection: &collection,
		Hostname:   "localhost",
		Port:       "3002",
	}

	return &tester, nil
}

// TestRequestWithName finds the named request in the collection, makes the same request, and then returns the request, expected response, and actual response
func (t *Tester) TestRequestWithName(name string) (*postman.Request, *postman.Response, *postman.Response, error) {
	itm := t.Collection.ItemWithName(name)
	if itm == nil {
		return nil, nil, nil, fmt.Errorf("item with name %s doesn't exist", name)
	}

	httpReq := itm.Request.ToHTTPRequest()
	if httpReq == nil {
		return nil, nil, nil, errors.New("failed to build HTTP request")
	}

	httpReq.URL.Host = fmt.Sprintf("%s:%s", t.Hostname, t.Port)
	httpReq.URL.Scheme = "http"

	resp, err := t.Client.Do(httpReq)
	if err != nil {
		return nil, nil, nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, nil, err
	}

	actual := &postman.Response{
		Mode:   "raw",
		Raw:    string(body),
		Status: resp.StatusCode,
	}

	return &itm.Request, &itm.Response[0], actual, nil
}
