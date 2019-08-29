// Package gzip provider a nice middleware for compress to responses.
package middleware

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	nice "../"
)

// Options represents a struct for specifying configuration options for the GZip middleware.
type Options struct {
	// Compression level. Can be DefaultCompression(-1), ConstantCompression(-2)
	// or any integer value between BestSpeed(1) and BestCompression(9) inclusive.
	CompressionLevel int
}

type gzipResponseWriter struct {
	w *gzip.Writer
	c *nice.Context
}

const (
	HEADER_ACCEPT_ENCODING  = "Accept-Encoding"
	HEADER_CONTENT_ENCODING = "Content-Encoding"
	HEADER_CONTENT_LENGTH   = "Content-Length"
	HEADER_CONTENT_TYPE     = "Content-Type"
	HEADER_VARY             = "Vary"
	SCHEME                  = "gzip"
)

// Gzip returns a nice middleware for compress to responses
func Gzip(opt Options) nice.HandlerFunc {
	pool := gzipPool(opt)
	return func(c *nice.Context) {
		w := c.Resp.GetWriter()
		if strings.Contains(c.Req.Header.Get(HEADER_ACCEPT_ENCODING), SCHEME) {
			c.Resp.Header().Add(HEADER_VARY, HEADER_ACCEPT_ENCODING)
			c.Resp.Header().Add(HEADER_CONTENT_ENCODING, SCHEME)
			gw := pool.Get().(*gzip.Writer)
			gw.Reset(w)
			defer func() {
				if c.Resp.Size() == 0 {
					// We have to reset response to it's pristine state when
					// nothing is written to body or error is returned.
					c.Resp.SetWriter(w)
					c.Resp.Header().Del(HEADER_CONTENT_ENCODING)
					gw.Reset(ioutil.Discard)
				}
				gw.Close()
				pool.Put(gw)
			}()
			c.Resp.SetWriter(gzipResponseWriter{gw, c})
		}

		c.Next()
	}
}

// Write writes the gzip data to the connection
func (grw gzipResponseWriter) Write(p []byte) (int, error) {
	if len(grw.c.Resp.Header().Get(HEADER_CONTENT_TYPE)) == 0 {
		grw.c.Resp.Header().Set(HEADER_CONTENT_TYPE, http.DetectContentType(p))
	}
	i, e := grw.w.Write(p)
	return i, e
}

func gzipPool(opt Options) sync.Pool {
	return sync.Pool{
		New: func() interface{} {
			w, _ := gzip.NewWriterLevel(ioutil.Discard, opt.CompressionLevel)
			return w
		},
	}
}
