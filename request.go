package gor

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type bodyData struct {
	JSON           interface{}
	FormURLEncoded map[string][]string
	FormData       map[string][]string
}

// Req is http Request struct
// <scheme>://<username>:<password>@<host>:<port>/<path>;<parameters>?<query>#<fragment>
type Req struct {
	r       *http.Request
	context context.Context

	Protocol string
	Secure   bool
	Method   string
	Query    map[string][]string
	Headers  map[string][]string
	Hostname string

	BaseURL     string
	OriginalURL string

	Params map[string]string
	Body   *bodyData
}

func getProtocol(r *http.Request) string {
	if r.TLS == nil {
		return "http"
	}

	return "https"
}

func getQuery(r *http.Request) (map[string][]string, error) {
	var rowQuery string
	if r.URL.RawQuery != "" {
		rowQuery = r.URL.RawQuery
	} else {
		URL, err := url.Parse(r.URL.Path)
		if err != nil {
			return nil, err
		}
		rowQuery = URL.RawQuery
	}

	query, err := url.ParseQuery(rowQuery)
	if err != nil {
		return nil, err
	}
	return query, nil
}

func getHostname(r *http.Request) string {
	hostPort := strings.Split(r.Host, ":")
	if len(hostPort) > 0 {
		return hostPort[0]
	}
	return ""
}

func getBaseURL(r *http.Request) string {
	return strings.Split(r.URL.Path, "?")[0]
}

func getOriginalURL(r *http.Request) string {
	return r.URL.Path
}

func getBody(r *http.Request) (*bodyData, error) {
	var t interface{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if err = json.Unmarshal(body, &t); err == nil {
		return &bodyData{JSON: t}, nil
	}

	if err := r.ParseForm(); err == nil && r.PostForm.Encode() != "" {
		return &bodyData{FormURLEncoded: r.PostForm}, nil
	}

	if err := r.ParseMultipartForm(2 ^ 10); err == nil {
		return &bodyData{FormData: r.PostForm}, nil
	}

	return nil, nil
}

func httpRequestToReq(r *http.Request) (*Req, error) {
	query, err := getQuery(r)
	if err != nil {
		return nil, err
	}

	body, err := getBody(r)
	if err != nil {
		return nil, err
	}

	protocol := getProtocol(r)

	return &Req{
		r:       r,
		context: r.Context(),

		Protocol: protocol,
		Secure:   protocol == "https",
		Method:   r.Method,
		Query:    query,
		Headers:  r.Header,
		Hostname: getHostname(r),

		BaseURL:     getBaseURL(r),
		OriginalURL: getOriginalURL(r),

		Params: make(map[string]string),
		Body:   body,
	}, nil
}

// AddContext add value to gor context
func (req *Req) AddContext(key, val interface{}) {
	req.context = context.WithValue(req.context, key, val)
}

// GetContext get context from gor by key
func (req *Req) GetContext(key interface{}) interface{} {
	return req.context.Value(key)
}

// BindJSON body to json
func (req *Req) BindJSON(v interface{}) error {
	defer io.Copy(ioutil.Discard, req.r.Body)
	return json.NewDecoder(req.r.Body).Decode(v)
}
