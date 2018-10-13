// Package recovery provider a nice middleware which recovers from panics anywhere in the chain.
package middleware

import (
	"fmt"
	"runtime"

	nice "../"
)

// Recovery returns a nice middleware which recovers from panics anywhere in the chain
// and handles the control to the centralized HTTPErrorHandler.
func Recovery() nice.HandlerFunc {
	return func(c *nice.Context) {
		defer func() {
			if err := recover(); err != nil {
				trace := make([]byte, 1<<16)
				n := runtime.Stack(trace, true)
				c.Error(fmt.Errorf("panic recover\n %v\n stack trace %d bytes\n %s", err, n, trace[:n]))
			}
		}()

		c.Next()
	}
}