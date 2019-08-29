// Package accesslog provider a nice middleware for log http access.
package middleware

import (
	"time"

	nice "../"
)

// Logger returns a nice middleware for log http access
func Logger() nice.HandlerFunc {
	return func(c *nice.Context) {
		start := time.Now()

		c.Next()

		c.Nice().Logger().Printf("%s %s %s %v %v", c.RemoteAddr(), c.Req.Method, c.URL(false), c.Resp.Status(), time.Since(start))
	}
}
