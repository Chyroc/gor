package gor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/unrolled/render"
)

func httpResponseWriterToRes(httpResponseWriter http.ResponseWriter, g *Gor) *Res {
	return &Res{
		httpResponseWriter,
		false,
		render.New(render.Options{Directory: g.renderDir}),
		nil,
		200,
	}
}

// Res is http ResponseWriter and some gor Response method
type Res struct {
	w      http.ResponseWriter
	exit   bool
	render *render.Render

	Response   interface{}
	StatusCode int
}

// Status set Response http status code
func (res *Res) Status(code int) *Res {
	res.StatusCode = code
	if http.StatusText(code) == "" {
		res.Status(http.StatusInternalServerError).Send(ErrHTTPStatusCodeInvalid)
	}

	return res
}

// SendStatus set Response http status code with its text
func (res *Res) SendStatus(code int) {
	if res.exit {
		return
	}

	res.Status(code).Send(http.StatusText(code))
}

func (res *Res) Write(data []byte) (int, error) {
	res.exit = true
	res.Response = string(data)
	res.w.WriteHeader(res.StatusCode)
	return res.w.Write(data)
}

// Send Send a Response
func (res *Res) Send(v interface{}) {
	if res.exit {
		return
	}

	res.Write([]byte(fmt.Sprintf("%v", v)))
	res.exit = true
}

// JSON Send json Response
func (res *Res) JSON(v interface{}) {
	defer func() {
		res.exit = true
	}()

	if res.exit {
		return
	}

	if v == nil {
		res.w.Header().Set("Content-Type", "application/json")
		res.Write([]byte("null"))
		return
	}

	t := reflect.TypeOf(v)
	switch t.Kind() {
	case reflect.Map, reflect.Struct, reflect.Slice, reflect.Array:
		break
	default:
		res.Status(http.StatusInternalServerError).Send(fmt.Sprintf("[%s] [%s] %+v", ErrResponseTypeUnsupported, t.Kind().String(), v))
		return
	}

	b, err := json.Marshal(v)
	if err != nil {
		res.Status(http.StatusInternalServerError).Send(ErrJSONMarshal)
		return
	}

	res.w.Header().Set("Content-Type", "application/json")
	res.Write(b)
}

// JSONP Send json Response with JSONP support
func (res *Res) JSONP(v interface{}) {
	// todo
}

// HTML render HTML
func (res *Res) HTML(v string, data interface{}) {
	if err := res.render.HTML(res, res.StatusCode, v, data); err != nil {
		res.Error(err.Error())
		return
	}
	res.exit = true
}

// Redirect Redirect to another url
func (res *Res) Redirect(path string) {
	res.w.Header().Set("Location", path)
	res.Status(http.StatusFound).Write([]byte(fmt.Sprintf(`%s. Redirecting to %s`, http.StatusText(http.StatusFound), path)))
}

// Download download
func (res *Res) Download(filepath string, filename ...string) {
	// todo
}

// Error send erroe Response
func (res *Res) Error(v string) {
	res.Status(http.StatusInternalServerError).Send(v)
}

// End end the request
func (res *Res) End() {
	res.exit = true
}

// AddHeader append (key, val) to headers
func (res *Res) AddHeader(key, val string) {
	res.w.Header().Add(key, val)
}

// SetCookie set cookie
func (res *Res) SetCookie(key, val string, option ...Cookie) {
	var cookie *http.Cookie
	if len(option) > 1 {
		res.Error("only support one cookie option")
	} else if len(option) == 1 {
		cookie = option[0].toHTTPCookie(key, val)
	} else {
		cookie = &http.Cookie{
			Name:  key,
			Value: val,
		}
	}

	http.SetCookie(res.w, cookie)
}

// ClearCookie clear cookie
func (res *Res) ClearCookie(key string, option ...Cookie) {
	// todo
}
