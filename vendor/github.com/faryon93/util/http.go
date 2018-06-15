package util

// util
// Copyright (C) 2018 Maximilian Pachl

// MIT License
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// ---------------------------------------------------------------------------------------
//  imports
// ---------------------------------------------------------------------------------------

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/schema"
)

// ---------------------------------------------------------------------------------------
//  errors
// ---------------------------------------------------------------------------------------

var (
	ErrInvalidContentType = errors.New("invalid content type")
)

// ---------------------------------------------------------------------------------------
//  types
// ---------------------------------------------------------------------------------------

// RequestBody represents the body of a request
// which was saved with SaveRequestBody().
type RequestBody struct {
	*bytes.Buffer
}

// ---------------------------------------------------------------------------------------
//  public functions
// ---------------------------------------------------------------------------------------

// GetRemoteAddr returns the remote address of an http request.
// If an X-Forwarded-For Header is present the headers content is returned.
// Otherwise the src host of the ip packet is returned.
func GetRemoteAddr(r *http.Request) string {
	remote := r.RemoteAddr
	if fwd := r.Header.Get("X-Forwarded-For"); len(fwd) > 0 {
		return fwd
	}

	// remove the port from the remote address
	if host, _, err := net.SplitHostPort(remote); err == nil {
		remote = host
	}

	return remote
}

// GetHostname return the Host Header of an http request.
// The Host is stripped from any port.
func GetHostname(r *http.Request) string {
	return StripPort(r.Host)
}

// Jsonify writes the JSON representation of v to the supplied
// http.ResposeWriter. If an error occours while marshalling the
// http response will be an internal server error.
func Jsonify(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// ParseBody reads the body of the request and parses it into v.
func ParseBody(r *http.Request, v interface{}) error {
	header := strings.Split(r.Header.Get("Content-Type"), ";")
	if len(header) < 1 {
		return ErrInvalidContentType
	}

	switch header[0] {
	case "application/x-www-form-urlencoded":
		err := r.ParseForm()
		if err != nil {
			return err
		}

		decoder := schema.NewDecoder()
		decoder.IgnoreUnknownKeys(true)
		return decoder.Decode(v, r.Form)

	default:
		return json.NewDecoder(r.Body).Decode(v)
	}
}

// SaveRequestBody saves the body of a request in order
// to restore it later.
func SaveRequestBody(r *http.Request) *RequestBody {
	// read the whole body
	body, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	// reinsert the new body into the request in order to be read again
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	return &RequestBody{bytes.NewBuffer(body)}
}

// Restore reinserts the saved body into the given request.
func (b *RequestBody) Restore(r *http.Request) *http.Request {
	r.Body = ioutil.NopCloser(b)
	return r
}
