package gor

import (
	"net/http"
)

type gorInterface interface {
	http.Handler
}

type resInterface interface {
	Write(data []byte) (int, error)
	Status(code int) *Res
	SendStatus(code int)
	Send(v interface{})
	JSON(v interface{})
	Redirect(path string)
	AddHeader(key, val string)
	SetCookie(key, val string, option ...Cookie)
	Error(v string)
	End()
}

type reqInterface interface {
	AddContext(key, val interface{})
	GetContext(key interface{}) interface{}
}

type normalMethod interface {
	Get(pattern string, h HandlerFunc)
	Head(pattern string, h HandlerFunc)
	Post(pattern string, h HandlerFunc)
	Put(pattern string, h HandlerFunc)
	Patch(pattern string, h HandlerFunc)
	Delete(pattern string, h HandlerFunc)
	Connect(pattern string, h HandlerFunc)
	Options(pattern string, h HandlerFunc)
	Trace(pattern string, h HandlerFunc)
}

// Mid mid
type Mid interface {
	handler(pattern string) *Route
}

// RouteInterface define Route Interface
type RouteInterface interface {
	Use(h HandlerFuncDefer)
	UseN(pattern string, m Mid)

	normalMethod
	Mid
}

var _ gorInterface = (*Gor)(nil)
var _ resInterface = (*Res)(nil)
var _ reqInterface = (*Req)(nil)
var _ RouteInterface = (*Gor)(nil)
var _ RouteInterface = (*Route)(nil)
