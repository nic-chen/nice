package nice

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRequest1(t *testing.T) {
	Convey("request body", t, func() {

		n.Get("/body", func(c *Context) {
			s, err := c.Body().String()
			So(err, ShouldBeNil)
			So(s, ShouldEqual, "{}")

			// because one req can only read once
			n, err := c.Body().Bytes()
			So(err, ShouldBeNil)
			So(string(n), ShouldEqual, "")

			r := c.Body().ReadCloser()
			n, err = ioutil.ReadAll(r)
			So(err, ShouldBeNil)
			So(string(n), ShouldEqual, "")
		})
		body := bytes.NewBuffer(nil)
		body.Write([]byte("{}"))
		req, _ := http.NewRequest("GET", "/body", body)
		w := httptest.NewRecorder()
		n.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}
