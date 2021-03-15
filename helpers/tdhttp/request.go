// Copyright (c) 2019, Maxime Soulé
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package tdhttp

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/maxatome/go-testdeep/internal/color"
	"github.com/maxatome/go-testdeep/internal/dark"
	"github.com/maxatome/go-testdeep/internal/flat"
)

func addHeaders(req *http.Request, headers []interface{}) (*http.Request, error) {
	headers = flat.Interfaces(headers...)

	for i := 0; i < len(headers); i++ {
		switch cur := headers[i].(type) {
		case string:
			i++
			var val string
			if i < len(headers) {
				var ok bool
				if val, ok = headers[i].(string); !ok {
					return nil, errors.New(color.Bad(
						`header "%s" should have a string value, not a %T (@ headers[%d])`,
						cur, headers[i], i))
				}
			}
			req.Header.Add(cur, val)

		case http.Header:
			for k, v := range cur {
				req.Header[k] = append(req.Header[k], v...)
			}

		default:
			return nil, errors.New(color.Bad(
				"headers... can only contains string and http.Header, not %T (@ headers[%d])",
				cur, i))
		}
	}
	return req, nil
}

// NewRequest creates a new HTTP request as
// net/http/httptest.NewRequest does, with the ability to immediately
// add some headers using string pairs as in:
//
//   req := NewRequest("POST", "/pdf", body,
//     "Content-type", "application/pdf",
//     "X-Test", "value",
//   )
//
// or using http.Header as in:
//
//   req := NewRequest("POST", "/pdf", body,
//     http.Header{"Content-type": []string{"application/pdf"}},
//   )
//
// Several header sources are combined:
//
//   req := NewRequest("POST", "/pdf", body,
//     "Content-type", "application/pdf",
//     http.Header{"X-Test": []string{"value1"}},
//     "X-Test", "value2",
//   )
//
// Produce the following http.Header:
//
//   http.Header{
//     "Content-type": []string{"application/pdf"},
//     "X-Test":       []string{"value1", "value2"},
//   }
//
// A string slice or a map can be flatened as well. As NewRequest() expects
// ...interface{}, td.Flatten() can help here too:
//   strHeaders := map[string]string{
//     "X-Length": "666",
//     "X-Foo":    "bar",
//   }
//   req := NewRequest("POST", "/pdf", body, td.Flatten(strHeaders))
//
// Or combined with forms seen above:
//   req := NewRequest("POST", "/pdf",
//     "Content-type", "application/pdf",
//     http.Header{"X-Test": []string{"value1"}},
//     td.Flatten(strHeaders),
//     "X-Test", "value2",
//   )
func NewRequest(method, target string, body io.Reader, headers ...interface{}) *http.Request {
	req, err := addHeaders(httptest.NewRequest(method, target, body), headers)
	if err != nil {
		f := dark.GetFatalizer()
		f.Helper()
		f.Fatal(err)
	}
	return req
}

// Get creates a new HTTP GET. It is a shortcut for:
//
//   NewRequest(http.MethodGet, target, nil, headers...)
//
// See NewRequest for all possible formats accepted in headers.
func Get(target string, headers ...interface{}) *http.Request {
	req, err := addHeaders(httptest.NewRequest(http.MethodGet, target, nil), headers)
	if err != nil {
		f := dark.GetFatalizer()
		f.Helper()
		f.Fatal(err)
	}
	return req
}

// Head creates a new HTTP HEAD. It is a shortcut for:
//
//   NewRequest(http.MethodHead, target, nil, headers...)
//
// See NewRequest for all possible formats accepted in headers.
func Head(target string, headers ...interface{}) *http.Request {
	req, err := addHeaders(httptest.NewRequest(http.MethodHead, target, nil), headers)
	if err != nil {
		f := dark.GetFatalizer()
		f.Helper()
		f.Fatal(err)
	}
	return req
}

// Post creates a HTTP POST. It is a shortcut for:
//
//   NewRequest(http.MethodPost, target, body, headers...)
//
// See NewRequest for all possible formats accepted in headers.
func Post(target string, body io.Reader, headers ...interface{}) *http.Request {
	req, err := addHeaders(httptest.NewRequest(http.MethodPost, target, body), headers)
	if err != nil {
		f := dark.GetFatalizer()
		f.Helper()
		f.Fatal(err)
	}
	return req
}

// PostForm creates a HTTP POST with data's keys and values
// URL-encoded as the request body. "Content-Type" header is
// automatically set to "application/x-www-form-urlencoded". Other
// headers can be added via headers, as in:
//
//   req := PostForm("/data",
//     url.Values{
//       "param1": []string{"val1", "val2"},
//       "param2": []string{"zip"},
//     },
//     "X-Foo", "Foo-value",
//     "X-Zip", "Zip-value",
//   )
//
// See NewRequest for all possible formats accepted in headers.
func PostForm(target string, data url.Values, headers ...interface{}) *http.Request {
	req, err := addHeaders(
		httptest.NewRequest(http.MethodPost, target, strings.NewReader(data.Encode())),
		append(headers, "Content-Type", "application/x-www-form-urlencoded"))
	if err != nil {
		f := dark.GetFatalizer()
		f.Helper()
		f.Fatal(err)
	}
	return req
}

// Put creates a HTTP PUT. It is a shortcut for:
//
//   NewRequest(http.MethodPut, target, body, headers...)
//
// See NewRequest for all possible formats accepted in headers.
func Put(target string, body io.Reader, headers ...interface{}) *http.Request {
	req, err := addHeaders(httptest.NewRequest(http.MethodPut, target, body), headers)
	if err != nil {
		f := dark.GetFatalizer()
		f.Helper()
		f.Fatal(err)
	}
	return req
}

// Patch creates a HTTP PATCH. It is a shortcut for:
//
//   NewRequest(http.MethodPatch, target, body, headers...)
//
// See NewRequest for all possible formats accepted in headers.
func Patch(target string, body io.Reader, headers ...interface{}) *http.Request {
	req, err := addHeaders(httptest.NewRequest(http.MethodPatch, target, body), headers)
	if err != nil {
		f := dark.GetFatalizer()
		f.Helper()
		f.Fatal(err)
	}
	return req
}

// Delete creates a HTTP DELETE. It is a shortcut for:
//
//   NewRequest(http.MethodDelete, target, body, headers...)
//
// See NewRequest for all possible formats accepted in headers.
func Delete(target string, body io.Reader, headers ...interface{}) *http.Request {
	req, err := addHeaders(httptest.NewRequest(http.MethodDelete, target, body), headers)
	if err != nil {
		f := dark.GetFatalizer()
		f.Helper()
		f.Fatal(err)
	}
	return req
}

func newJSONRequest(method, target string, body interface{}, headers ...interface{}) (*http.Request, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, errors.New(color.Bad("JSON encoding failed: %s", err))
	}

	return addHeaders(NewRequest(method, target, bytes.NewBuffer(b)),
		append(headers[:len(headers):len(headers)],
			"Content-Type", "application/json"))
}

// NewJSONRequest creates a new HTTP request with body marshaled to
// JSON. "Content-Type" header is automatically set to
// "application/json". Other headers can be added via headers, as in:
//
//   req := NewJSONRequest("POST", "/data", body,
//     "X-Foo", "Foo-value",
//     "X-Zip", "Zip-value",
//   )
//
// See NewRequest for all possible formats accepted in headers.
func NewJSONRequest(method, target string, body interface{}, headers ...interface{}) *http.Request {
	req, err := newJSONRequest(method, target, body, headers...)
	if err != nil {
		f := dark.GetFatalizer()
		f.Helper()
		f.Fatal(err)
	}
	return req
}

// PostJSON creates a HTTP POST with body marshaled to
// JSON. "Content-Type" header is automatically set to
// "application/json". It is a shortcut for:
//
//   NewJSONRequest(http.MethodPost, target, body, headers...)
//
// See NewRequest for all possible formats accepted in headers.
func PostJSON(target string, body interface{}, headers ...interface{}) *http.Request {
	req, err := newJSONRequest(http.MethodPost, target, body, headers...)
	if err != nil {
		f := dark.GetFatalizer()
		f.Helper()
		f.Fatal(err)
	}
	return req
}

// PutJSON creates a HTTP PUT with body marshaled to
// JSON. "Content-Type" header is automatically set to
// "application/json". It is a shortcut for:
//
//   NewJSONRequest(http.MethodPut, target, body, headers...)
//
// See NewRequest for all possible formats accepted in headers.
func PutJSON(target string, body interface{}, headers ...interface{}) *http.Request {
	req, err := newJSONRequest(http.MethodPut, target, body, headers...)
	if err != nil {
		f := dark.GetFatalizer()
		f.Helper()
		f.Fatal(err)
	}
	return req
}

// PatchJSON creates a HTTP PATCH with body marshaled to
// JSON. "Content-Type" header is automatically set to
// "application/json". It is a shortcut for:
//
//   NewJSONRequest(http.MethodPatch, target, body, headers...)
//
// See NewRequest for all possible formats accepted in headers.
func PatchJSON(target string, body interface{}, headers ...interface{}) *http.Request {
	req, err := newJSONRequest(http.MethodPatch, target, body, headers...)
	if err != nil {
		f := dark.GetFatalizer()
		f.Helper()
		f.Fatal(err)
	}
	return req
}

// DeleteJSON creates a HTTP DELETE with body marshaled to
// JSON. "Content-Type" header is automatically set to
// "application/json". It is a shortcut for:
//
//   NewJSONRequest(http.MethodDelete, target, body, headers...)
//
// See NewRequest for all possible formats accepted in headers.
func DeleteJSON(target string, body interface{}, headers ...interface{}) *http.Request {
	req, err := newJSONRequest(http.MethodDelete, target, body, headers...)
	if err != nil {
		f := dark.GetFatalizer()
		f.Helper()
		f.Fatal(err)
	}
	return req
}

func newXMLRequest(method, target string, body interface{}, headers ...interface{}) (*http.Request, error) {
	b, err := xml.Marshal(body)
	if err != nil {
		return nil, errors.New(color.Bad("XML encoding failed: %s", err))
	}

	return addHeaders(NewRequest(method, target, bytes.NewBuffer(b)),
		append(headers[:len(headers):len(headers)],
			"Content-Type", "application/xml"))
}

// NewXMLRequest creates a new HTTP request with body marshaled to
// XML. "Content-Type" header is automatically set to
// "application/xml". Other headers can be added via headers, as in:
//
//   req := NewXMLRequest("POST", "/data", body,
//     "X-Foo", "Foo-value",
//     "X-Zip", "Zip-value",
//   )
//
// See NewRequest for all possible formats accepted in headers.
func NewXMLRequest(method, target string, body interface{}, headers ...interface{}) *http.Request {
	req, err := newXMLRequest(method, target, body, headers...)
	if err != nil {
		f := dark.GetFatalizer()
		f.Helper()
		f.Fatal(err)
	}
	return req
}

// PostXML creates a HTTP POST with body marshaled to
// XML. "Content-Type" header is automatically set to
// "application/xml". It is a shortcut for:
//
//   NewXMLRequest(http.MethodPost, target, body, headers...)
//
// See NewRequest for all possible formats accepted in headers.
func PostXML(target string, body interface{}, headers ...interface{}) *http.Request {
	req, err := newXMLRequest(http.MethodPost, target, body, headers...)
	if err != nil {
		f := dark.GetFatalizer()
		f.Helper()
		f.Fatal(err)
	}
	return req
}

// PutXML creates a HTTP PUT with body marshaled to
// XML. "Content-Type" header is automatically set to
// "application/xml". It is a shortcut for:
//
//   NewXMLRequest(http.MethodPut, target, body, headers...)
//
// See NewRequest for all possible formats accepted in headers.
func PutXML(target string, body interface{}, headers ...interface{}) *http.Request {
	req, err := newXMLRequest(http.MethodPut, target, body, headers...)
	if err != nil {
		f := dark.GetFatalizer()
		f.Helper()
		f.Fatal(err)
	}
	return req
}

// PatchXML creates a HTTP PATCH with body marshaled to
// XML. "Content-Type" header is automatically set to
// "application/xml". It is a shortcut for:
//
//   NewXMLRequest(http.MethodPatch, target, body, headers...)
//
// See NewRequest for all possible formats accepted in headers.
func PatchXML(target string, body interface{}, headers ...interface{}) *http.Request {
	req, err := newXMLRequest(http.MethodPatch, target, body, headers...)
	if err != nil {
		f := dark.GetFatalizer()
		f.Helper()
		f.Fatal(err)
	}
	return req
}

// DeleteXML creates a HTTP DELETE with body marshaled to
// XML. "Content-Type" header is automatically set to
// "application/xml". It is a shortcut for:
//
//   NewXMLRequest(http.MethodDelete, target, body, headers...)
//
// See NewRequest for all possible formats accepted in headers.
func DeleteXML(target string, body interface{}, headers ...interface{}) *http.Request {
	req, err := newXMLRequest(http.MethodDelete, target, body, headers...)
	if err != nil {
		f := dark.GetFatalizer()
		f.Helper()
		f.Fatal(err)
	}
	return req
}
