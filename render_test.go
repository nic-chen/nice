package nice

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRender1(t *testing.T) {
	Convey("response method", t, func() {
		n.Get("/render", func(c *Context) {
			c.Set("name", "Nice")
			c.HTML(200, "_fixture/index1.html")
		})

		n.Get("/render2", func(c *Context) {
			c.Set("name", "Nice")
			c.HTML(200, "_fixture/index2.html")
		})

		n.Get("/render3", func(c *Context) {
			c.Set("name", "Nice")
			c.HTML(200, "_fixture/index3.html")
		})

		req, _ := http.NewRequest("GET", "/render", nil)
		w := httptest.NewRecorder()
		n.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, http.StatusOK)

		req, _ = http.NewRequest("GET", "/render2", nil)
		w = httptest.NewRecorder()
		n.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		req, _ = http.NewRequest("GET", "/render3", nil)
		w = httptest.NewRecorder()
		n.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})
}
