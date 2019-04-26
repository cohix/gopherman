package gopherman

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"

	"github.com/cohix/gopherman/postman"

	"github.com/stretchr/testify/assert"
)

// Tester represents a collection test tool
type Tester struct {
	Environment *postman.Environment
	Client      *http.Client
	Collections []postman.Collection
	Hostname    string
	Port        string
}

// NewTesterWithCollection loads a collection from a file
func NewTesterWithCollection(path string, envFile string, files ...string) (*Tester, error) {
	env, err := postman.EnvironmentFromFile(filepath.Join(path, envFile))
	if err != nil {
		return nil, err
	}

	collections := make([]postman.Collection, len(files))

	for i, name := range files {
		file, err := ioutil.ReadFile(filepath.Join(path, name))
		if err != nil {
			return nil, err
		}

		collection := postman.Collection{}
		if err := json.Unmarshal(file, &collection); err != nil {
			return nil, err
		}

		collections[i] = collection
	}

	tester := Tester{
		Environment: env,
		Client:      http.DefaultClient,
		Collections: collections,
		Hostname:    "localhost",
		Port:        "3002",
	}

	return &tester, nil
}

// TestRequestWithName finds the named request in the collection, makes the same request, and then returns the request, expected response, and actual response
func (t *Tester) TestRequestWithName(name string, tst *testing.T, handler func(*TestHelper, *postman.Request, *postman.Response, *postman.Response)) []error {
	vars := t.Environment.VariableMap()

	tmplHost, err := postman.SubstVars("{{ .BaseUrl }}:{{ .Port }}", vars)
	if err != nil {
		fmt.Println(err)
		tmplHost = "localhost:8080"
	}

	errs := []error{}

	for _, collection := range t.Collections {
		helper := NewTestHelper(tst)

		// put this in a func so that critical errors can be collected and then bail out
		func() {
			itm := collection.ItemWithName(name)
			if itm == nil {
				helper.Error(fmt.Errorf("item with name %s doesn't exist", name))
				return
			}

			httpReq := itm.Request.ToHTTPRequest(vars)
			if httpReq == nil {
				helper.Error(errors.New("failed to build HTTP request"))
				return
			}

			httpReq.URL.Host = tmplHost
			httpReq.URL.Scheme = "http"

			actual, err := makeRequest(t.Client, httpReq)
			if err != nil {
				helper.Error(err)
				return
			}

			handler(helper, &itm.Request, &itm.Response[0], actual)
		}()

		if len(helper.Errors) > 0 {
			for _, e := range helper.Errors {
				wrapped := errors.Wrapf(e, "(collection %s, request %s)", collection.Info.Name, name)
				errs = append(errs, wrapped)
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func makeRequest(client *http.Client, req *http.Request) (*postman.Response, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("request returned error statusCode: %d %s", resp.StatusCode, string(body))
	}

	actual := &postman.Response{
		Mode:   "raw",
		Raw:    string(body),
		Status: resp.StatusCode,
	}

	return actual, nil
}

// TestHelper helps with running tests
type TestHelper struct {
	Assert *assert.Assertions
	Errors []error
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	helper := &TestHelper{
		Assert: assert.New(t),
		Errors: []error{},
	}

	return helper
}

func (t *TestHelper) Error(err error) {
	t.Errors = append(t.Errors, err)
}

// Log logs something
func (t *TestHelper) Log(msg string) {
	fmt.Println(msg)
}
