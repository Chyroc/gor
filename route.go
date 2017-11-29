package gor

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

// HandlerFunc gor handler func like http.HandlerFunc func(ResponseWriter, *Request)
type HandlerFunc func(*Req, *Res)

// Next exec next handler or mid
type Next func()

// HandlerFuncNext gor handler func like http.HandlerFunc func(ResponseWriter, *Request),
// but return HandlerFunc to do somrthing at defer time
type HandlerFuncNext func(*Req, *Res, Next)

type routeParam struct {
	name    string
	isParam bool
}

type matchType int

const (
	preMatch  matchType = iota
	fullMatch
)

type route struct {
	method    string
	routePath string
	matchType matchType
	//prepath     string

	//routeParams  []*routeParam
	routePathReg *regexp.Regexp

	//parentIndex string
	handlerFunc     HandlerFunc
	handlerFuncNext HandlerFuncNext
	middleware      Middleware

	children []*route
}

type matchedRoute struct {
	index  int
	params map[string]string
}

func (r *route) copy() *route {
	var t = &route{
		method:    r.method,
		routePath: r.routePath,
		matchType: r.matchType,
		//prepath:   r.prepath,

		routePathReg: r.routePathReg,

		handlerFunc:     r.handlerFunc,
		handlerFuncNext: r.handlerFuncNext,
		middleware:      r.middleware,

		children: copyRouteSlice(r.children),
	}
	//var rs []*routeParam
	//for _, v := range r.routeParams {
	//	rs = append(rs, &routeParam{
	//		name:    v.name,
	//		isParam: v.isParam,
	//	})
	//}
	//t.routeParams = rs
	return t
}

func copyRouteSlice(routes []*route) []*route {
	var rs []*route
	for _, v := range routes {
		rs = append(rs, v.copy())
	}
	return rs
}

func mergeRouteParamSlice(parentRouteParams []*routeParam, subRouteParams ...*routeParam) []*routeParam {
	var rs []*routeParam
	for _, v := range parentRouteParams {
		rs = append(rs, &routeParam{
			name:    v.name,
			isParam: v.isParam,
		})
	}
	for _, v := range subRouteParams {
		rs = append(rs, &routeParam{
			name:    v.name,
			isParam: v.isParam,
		})
	}
	return rs
}

//[]*routeParam
// Route route
type Route struct {
	routes []*route
	mids   []HandlerFuncNext
}

// NewRoute return *Router
func NewRoute() *Route {
	return &Route{}
}

func (r *Route) addHandlerFuncAndNextRoute(method string, pattern string, matchType matchType, h HandlerFunc, hn HandlerFuncNext, parentrouteParams []*routeParam) {
	if !strings.HasPrefix(pattern, "/") {
		panic("must start with /")
	}
	if strings.HasSuffix(pattern, "/") && pattern != "/" {
		pattern = pattern[:len(pattern)-1]
	}

	URL, err := url.Parse(pattern)
	if err != nil {
		panic(fmt.Sprintf("pattern invalid: %s", pattern))
	}

	routePath := URL.Path

	//var rps []*routeParam
	//for _, i := range paths {
	//	if strings.HasPrefix(i, ":") {
	//		//parentrouteParams = append(parentrouteParams, &routeParam{name: i[1:], isParam: true})
	//		parentrouteParams = mergeRouteParamSlice(parentrouteParams, &routeParam{name: i[1:], isParam: true})
	//	} else {
	//		//parentrouteParams = append(parentrouteParams, &routeParam{name: i, isParam: false})
	//		parentrouteParams = mergeRouteParamSlice(parentrouteParams, &routeParam{name: i, isParam: false})
	//	}
	//}
	var routeH = &route{
		method:    method,
		routePath: routePath,
		matchType: matchType,

		routePathReg: genMatchPathReg(routePath),
	}
	if h != nil {
		routeH.handlerFunc = h
	} else if hn != nil {
		routeH.handlerFuncNext = hn
	} else {
		panic("handlerFunc or handlerFuncNext cannot be both nil")
	}

	r.routes = append(r.routes, routeH)
}

// Get http get method
func (r *Route) Get(pattern string, h HandlerFunc) {
	r.addHandlerFuncAndNextRoute(http.MethodGet, pattern, fullMatch, h, nil, []*routeParam{})
}

// Head http head method
func (r *Route) Head(pattern string, h HandlerFunc) {
	r.addHandlerFuncAndNextRoute(http.MethodHead, pattern, fullMatch, h, nil, []*routeParam{})
}

// Post http post method
func (r *Route) Post(pattern string, h HandlerFunc) {
	r.addHandlerFuncAndNextRoute(http.MethodPost, pattern, fullMatch, h, nil, []*routeParam{})
}

// Put http put method
func (r *Route) Put(pattern string, h HandlerFunc) {
	r.addHandlerFuncAndNextRoute(http.MethodPut, pattern, fullMatch, h, nil, []*routeParam{})
}

// Patch http patch method
func (r *Route) Patch(pattern string, h HandlerFunc) {
	r.addHandlerFuncAndNextRoute(http.MethodPatch, pattern, fullMatch, h, nil, []*routeParam{})
}

// Delete http delete method
func (r *Route) Delete(pattern string, h HandlerFunc) {
	r.addHandlerFuncAndNextRoute(http.MethodDelete, pattern, fullMatch, h, nil, []*routeParam{})
}

// Connect http connect method
func (r *Route) Connect(pattern string, h HandlerFunc) {
	r.addHandlerFuncAndNextRoute(http.MethodConnect, pattern, fullMatch, h, nil, []*routeParam{})
}

// Options http options method
func (r *Route) Options(pattern string, h HandlerFunc) {
	r.addHandlerFuncAndNextRoute(http.MethodOptions, pattern, fullMatch, h, nil, []*routeParam{})
}

// Trace http trace method
func (r *Route) Trace(pattern string, h HandlerFunc) {
	r.addHandlerFuncAndNextRoute(http.MethodTrace, pattern, fullMatch, h, nil, []*routeParam{})
}

// Use http trace method
//
// must belong one of below type
// string
// type HandlerFunc func(*Req, *Res)
// type HandlerFuncNext func(*Req, *Res, Next)
// type Middleware interface
func (r *Route) Use(hs ...interface{}) {
	if len(hs) == 1 {
		r.useWithOne("/", hs[0])
		return
	}

	first := hs[0]
	firstType := reflect.TypeOf(first)
	if firstType.Kind() == reflect.String {
		firstValue := reflect.ValueOf(first)
		pattern := firstValue.String()
		for _, h := range hs[1:] {
			r.useWithOne(pattern, h)
		}
	} else {
		for _, h := range hs {
			r.useWithOne("/", h)
		}
	}
}

func (r *Route) useWithOne(pattern string, h interface{}) {
	// todo use 应该处理签名的params
	var err error = nil
	defer func() {
		if err != nil {
			panic(err)
		}
	}()
	hType := reflect.TypeOf(h)
	switch hType.Kind() {
	case reflect.Func:
		switch h.(type) {
		case func(req *Req, res *Res):
			if f, ok := h.(func(req *Req, res *Res)); ok {
				r.useWithHandlerFunc("ALL", pattern, preMatch, HandlerFunc(f), []*routeParam{})
			} else {
				err = fmt.Errorf("cannot convert to gor.HandlerFunc")
			}
		case func(req *Req, res *Res, next Next):
			if f, ok := h.(func(req *Req, res *Res, next Next)); ok {
				r.useWithHandlerFuncNext("ALL", pattern, preMatch, HandlerFuncNext(f), []*routeParam{}) // todo parentrouteParams
			} else {
				err = fmt.Errorf("cannot convert to gor.HandlerFuncNext")
			}
		default:
			err = fmt.Errorf("maybe you are transmiting gor.HandlerFunc / gor.HandlerFuncNext, but the function signature is wrong")
		}
	case reflect.Struct:
		err = fmt.Errorf("maybe you are transmiting gor.Middleware, but please use Pointer, not Struct")
	case reflect.Ptr:
		// 添加子节点
		if f, ok := h.(Middleware); ok {
			r.useWithMiddleware("ALL", pattern, preMatch, Middleware(f), []*routeParam{}) // todo parentrouteParams
		} else {
			err = fmt.Errorf("cannot convert to gor.Middleware")
		}
	default:
		err = fmt.Errorf("when middleware length is one, that type must belong gor.HandlerFunc / gor.HandlerFuncNext / gor.Route, but get %s", hType.Kind())
	}
}

func (r *Route) handler(pattern string) []*route {
	a:=copyRouteSlice(r.routes)
	//fmt.Printf("\naaaaaa %s\n",a[0])
	//fmt.Printf("\naaaaaa %s\n",a[1])
	return a
	//return r.routes
}

func (r *Route) useWithHandlerFunc(method, pattern string, matchType matchType, h HandlerFunc, parentrouteParams []*routeParam) {
	//fmt.Printf("get pattern: %s, HandlerFunc: %+v\n", pattern, h)
	r.addHandlerFuncAndNextRoute(method, pattern, matchType, h, nil, parentrouteParams)
}

func (r *Route) useWithHandlerFuncNext(method, pattern string, matchType matchType, h HandlerFuncNext, parentrouteParams []*routeParam) {
	//fmt.Printf("get pattern: %s, HandlerFuncNext: %+v\n", pattern, h)
	r.addHandlerFuncAndNextRoute(method, pattern, matchType, nil, h, parentrouteParams)
}

func (r *Route) useWithMiddleware(method, pattern string, matchType matchType, mid Middleware, parentrouteParams []*routeParam) {
	if method != "ALL" {
		panic("middleware method must be ALL")
	}
	subRoutes := mid.handler(pattern)

	parent := &route{
		method:    method,
		matchType: matchType,
		routePath: pattern,
		//routeParams: nil, // todo

		//prepath:     pattern[1:],

		routePathReg: genMatchPathReg(pattern),

		children: subRoutes,
	}
	r.routes = append(r.routes, parent)
	//for _, subRoute := range subRoutes {
	//	fmt.Printf("subRoute %s\n", subRoute)
	//	var newParentrouteParams []*routeParam
	//	if subRoute.prepath != "" {
	//		fmt.Printf("1\n")
	//		if strings.HasPrefix(subRoute.prepath, ":") {
	//			fmt.Printf("2\n")
	//			newParentrouteParams = mergeRouteParamSlice(parentrouteParams, &routeParam{name: subRoute.prepath[1:], isParam: true})
	//		} else {
	//			fmt.Printf("3\n")
	//			newParentrouteParams = mergeRouteParamSlice(parentrouteParams, &routeParam{name: subRoute.prepath, isParam: false})
	//		}
	//		fmt.Printf("4\n")
	//	}
	//	fmt.Printf("5\n")
	//	newParentrouteParams = mergeRouteParamSlice(newParentrouteParams, subRoute.routeParams...)
	//	if subRoute.handlerFunc != nil {
	//		fmt.Printf("6\n")
	//		r.routes = append(r.routes, &route{
	//			handlerFunc: subRoute.handlerFunc,
	//			method:      subRoute.method,
	//			prepath:     pattern[1:],
	//			routeParams: newParentrouteParams,
	//		})
	//	} else if subRoute.handlerFuncNext != nil {
	//		fmt.Printf("7\n")
	//		fmt.Printf("6\n")
	//		r.routes = append(r.routes, &route{
	//			handlerFuncNext: subRoute.handlerFuncNext,
	//			method:          subRoute.method,
	//			prepath:         pattern[1:],
	//			routeParams:     newParentrouteParams,
	//		})
	//	} else if subRoute.middleware != nil {
	//		fmt.Printf("8\n")
	//		r.useWithMiddleware(subRoute.method, pattern+"/"+subRoute.prepath, subRoute.middleware, newParentrouteParams)
	//	} else {
	//		fmt.Printf("9\n")
	//		panic("notklsadjlfajs")
	//	}
	//}
}

//
//func (r *Route) useWithMiddleware(method, pattern string, mid Middleware, parentrouteParams []*routeParam) {
//	subRoutes := mid.handler(pattern)
//	for _, subRoute := range subRoutes {
//		fmt.Printf("subRoute %s\n", subRoute)
//		var newParentrouteParams []*routeParam
//		if subRoute.prepath != "" {
//			fmt.Printf("1\n")
//			if strings.HasPrefix(subRoute.prepath, ":") {
//				fmt.Printf("2\n")
//				newParentrouteParams = mergeRouteParamSlice(parentrouteParams, &routeParam{name: subRoute.prepath[1:], isParam: true})
//			} else {
//				fmt.Printf("3\n")
//				newParentrouteParams = mergeRouteParamSlice(parentrouteParams, &routeParam{name: subRoute.prepath, isParam: false})
//			}
//			fmt.Printf("4\n")
//		}
//		fmt.Printf("5\n")
//		newParentrouteParams = mergeRouteParamSlice(newParentrouteParams, subRoute.routeParams...)
//		if subRoute.handlerFunc != nil {
//			fmt.Printf("6\n")
//			r.routes = append(r.routes, &route{
//				handlerFunc: subRoute.handlerFunc,
//				method:      subRoute.method,
//				prepath:     pattern[1:],
//				routeParams: newParentrouteParams,
//			})
//		} else if subRoute.handlerFuncNext != nil {
//			fmt.Printf("7\n")
//			fmt.Printf("6\n")
//			r.routes = append(r.routes, &route{
//				handlerFuncNext: subRoute.handlerFuncNext,
//				method:          subRoute.method,
//				prepath:         pattern[1:],
//				routeParams:     newParentrouteParams,
//			})
//		} else if subRoute.middleware != nil {
//			fmt.Printf("8\n")
//			r.useWithMiddleware(subRoute.method, pattern+"/"+subRoute.prepath, subRoute.middleware, newParentrouteParams)
//		} else {
//			fmt.Printf("9\n")
//			panic("notklsadjlfajs")
//		}
//	}
//}
