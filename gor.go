package gor

import "fmt"

// HandlerFunc gor handler func like http.HandlerFunc func(ResponseWriter, *Request)
type HandlerFunc func(*Req, *Res)
type HandlerFuncDefer func(*Req, *Res) HandlerFunc

//func GenDeferFunc() *HandlerFunc{
//return &
//}

// Gor gor framework core struct
type Gor struct {
	*Route
}

// NewGor return Gor struct
func NewGor() *Gor {
	return &Gor{
		NewRoute(),
	}
}

var debug = false

func debugPrintf(format string, a ...interface{}) {
	if debug {
		fmt.Printf(format+"\n", a...)
	}
}
