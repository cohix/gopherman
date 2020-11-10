package gopherman

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/cohix/gopherman/postman"
	"github.com/pkg/errors"
)

// RequestRecorder allows requests to an http server to be recorded
type RequestRecorder struct {
	mux       http.Handler
	reqs      []postman.CollectionItem
	auth      *postman.CollectionAuth
	startTime *time.Time
}

func (rr *RequestRecorder) reset() {
	rr.reqs = []postman.CollectionItem{}
	now := time.Now()
	rr.startTime = &now
	return
}

// NewRequestRecorder returns a recorder ready to be used
func NewRequestRecorder(mux http.Handler) *RequestRecorder {
	rr := RequestRecorder{
		mux:  mux,
		reqs: []postman.CollectionItem{},
	}

	return &rr
}

func (rr *RequestRecorder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/gopherman-terminate" {
		rr.handleTerminate(w, r)
		return
	}

	if r.URL.Path == "/gopherman-reset" {
		rr.handleReset(w, r)
		return
	}

	if rr.startTime == nil {
		now := time.Now()
		rr.startTime = &now
	}

	req, err := postman.RequestFromHTTP(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to RequestFromHTTP: " + err.Error()))
		return
	}

	fakeWriter := NewFakeWriter(r.Header)

	rr.mux.ServeHTTP(fakeWriter, r)

	w.WriteHeader(fakeWriter.StatusCode)

	item := postman.CollectionItem{
		Name:    fmt.Sprintf("%s %s", r.Method, r.URL.RequestURI()),
		Request: *req,
	}

	if len(fakeWriter.Body) > 0 {
		if _, err := w.Write(fakeWriter.Body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		item.Response = []postman.Response{
			postman.Response{
				Mode:   "raw",
				Raw:    string(fakeWriter.Body),
				Status: fakeWriter.StatusCode,
			},
		}
	}

	rr.reqs = append(rr.reqs, item)

	fmt.Println("request recorded")
}

func (rr *RequestRecorder) handleTerminate(w http.ResponseWriter, r *http.Request) {
	fmt.Println("RequestRecorder terminating")
	if !rr.isStarted() {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("recorder is not started"))
		return
	}

	collection := postman.NewCollection(fmt.Sprintf("%s", rr.startTime), rr.reqs, rr.auth)

	collectionJSON, err := json.MarshalIndent(collection, "", "\t")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errors.Wrap(err, "failed to Marshal collection").Error()))
		return
	}

	filepath := filepathForSession(*rr.startTime)

	if err := ioutil.WriteFile(filepath, collectionJSON, 0700); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errors.Wrap(err, "failed to write collection").Error()))
		return
	}

	fmt.Printf("RequestRecorder wrote collection to %s\n", filepath)

	w.WriteHeader(http.StatusOK)
	w.Write(collectionJSON)
}

func (rr *RequestRecorder) handleReset(w http.ResponseWriter, r *http.Request) {
	fmt.Println("RequestRecorder resetting")
	if !rr.isStarted() {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("recorder is not started"))
		return
	}

	rr.reset()

	fmt.Println("RequestRecorder reset")

	return
}

func (rr *RequestRecorder) isStarted() bool {
	return rr.startTime != nil
}

func filepathForSession(start time.Time) string {
	dir := defaultFilePath()
	name := fmt.Sprintf("%s.json", start)

	return filepath.Join(dir, name)
}

func defaultFilePath() string {
	dir := filepath.Join(homeDir(), ".op", "gopherman")
	os.MkdirAll(dir, 0700)

	return dir
}

func homeDir() string {
	home := os.Getenv("HOME")
	if home != "" {
		return home
	}

	usr, err := user.Current()
	if err != nil {
		return ""
	}

	home = usr.HomeDir
	return home
}
