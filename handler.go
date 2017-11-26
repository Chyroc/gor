package gor

import (
	"net/http"
	"strings"
)

func splitRoute(r *http.Request) []string {
	return strings.Split(strings.Split(r.URL.Path[1:], "?")[0], "/")
}

func matchRouter(method string, paths []string, routes []*route) (map[string]string, int) {
	for _, v := range paths {
		if strings.Contains(v, "/") {
			panic("paths cannot contain /")
		}
	}
	matchIndex := -1
	for _, route := range routes {
		matchIndex++
		if route.prepath == paths[0] {
			if method != "ALL" && route.method != method {
				continue
			}
			matchRoutes := paths[1:]
			if len(matchRoutes) != len(route.routeParams) {
				continue
			}

			if len(route.routeParams) == 0 && len(matchRoutes) == 0 {
				return nil, matchIndex
			}

			match := false
			matchParams := make(map[string]string)
			for i, j := 0, len(matchRoutes); i < j; i++ {
				if route.routeParams[i].isParam {
					matchParams[route.routeParams[i].name] = matchRoutes[i]
				} else if route.routeParams[i].name != matchRoutes[i] {
					match = false
					break
				}
				match = true
			}
			if match {
				return matchParams, matchIndex
			}
		}
	}

	return nil, -1
}

// ServeHTTP use to start server
func (g *Gor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res := httpResponseWriterToRes(w)
	req, err := httpRequestToReq(r)
	if err != nil {
		res.Error(err.Error())
		return
	}

	routes := splitRoute(r)
	matchParams, matchIndex := matchRouter(r.Method, routes, g.routes)
	if matchIndex != -1 {
		if matchParams != nil {
			for k, v := range matchParams {
				req.Params[k] = v
			}
		}
		g.routes[matchIndex].handler(req, res)
		return
	}

	res.SendStatus(http.StatusNotFound)
}

// Use add middlewares
func (g *Gor) Use(middlewares ...func(g *Gor) http.Handler) {
	//g.middlewares = append(g.middlewares, middlewares...)
}

// Listen bind port and start server
func (g *Gor) Listen(addr string) error {
	return http.ListenAndServe(addr, g)
}
