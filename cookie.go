package gor

import (
	"net/http"
	"time"
)

// Cookie like http.Cookie but remove name and value
type Cookie struct {
	Path       string    // optional
	Domain     string    // optional
	Expires    time.Time // optional
	RawExpires string    // for reading cookies only

	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
	// MaxAge>0 means Max-Age attribute present and given in seconds
	MaxAge   int
	Secure   bool
	HTTPOnly bool
	Raw      string
	Unparsed []string // Raw text of unparsed attribute-value pairs
}

func (c *Cookie) toHTTPCookie(key string, val ...string) *http.Cookie {
	return &http.Cookie{
		Name:  key,
		Value: val[0],

		Path:       c.Path,
		Domain:     c.Domain,
		Expires:    c.Expires,
		RawExpires: c.RawExpires,

		MaxAge:   c.MaxAge,
		Secure:   c.Secure,
		HttpOnly: c.HTTPOnly,
		Raw:      c.Raw,
		Unparsed: c.Unparsed,
	}
}
