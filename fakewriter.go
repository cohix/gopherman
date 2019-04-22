package gopherman

import "net/http"

// FakeWriter represenrs a fake http.ResponseWriter
type FakeWriter struct {
	StatusCode int
	Body       []byte
	headers    http.Header
}

// NewFakeWriter returns a new FakeWriter
func NewFakeWriter(headers http.Header) *FakeWriter {
	return &FakeWriter{
		headers: headers,
	}
}

// WriteHeader writes the header value
func (w *FakeWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
}

// Write writes the response body
func (w *FakeWriter) Write(body []byte) (int, error) {
	w.Body = body

	if w.StatusCode == 0 {
		w.StatusCode = http.StatusOK
	}

	return len(body), nil
}

// Header returns the header
func (w *FakeWriter) Header() http.Header {
	return w.headers
}
