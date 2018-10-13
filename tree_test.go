package nice

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// print the route map
func (t *Tree) print(prefix string, root *leaf) {
	if root == nil {
		for m := range t.nodes {
			fmt.Println(m)
			t.print("", t.nodes[m])
		}
		return
	}
	prefix = fmt.Sprintf("%s -> %s", prefix, root.pattern)
	fmt.Println(prefix)
	root.children = append(root.children, root.paramChild)
	root.children = append(root.children, root.wideChild)
	for i := range root.children {
		if root.children[i] != nil {
			t.print(prefix, root.children[i])
		}
	}
}

func TestTreeRouteAdd1(t *testing.T) {
	Convey("add static route", t, func() {
		r.Add("GET", "/", []HandlerFunc{f})
		r.Add("GET", "/bcd", []HandlerFunc{f})
		r.Add("GET", "/abcd", []HandlerFunc{f})
		r.Add("GET", "/abc", []HandlerFunc{f})
		r.Add("GET", "/abd", []HandlerFunc{f})
		r.Add("GET", "/abcdef", []HandlerFunc{f})
		r.Add("GET", "/bcdefg", []HandlerFunc{f})
		r.Add("GET", "/abc/123", []HandlerFunc{f})
		r.Add("GET", "/abc/234", []HandlerFunc{f})
		r.Add("GET", "/abc/125", []HandlerFunc{f})
		r.Add("GET", "/abc/235", []HandlerFunc{f})
		r.Add("GET", "/cbd/123", []HandlerFunc{f})
		r.Add("GET", "/cbd/234", []HandlerFunc{f})
		r.Add("GET", "/cbd/345", []HandlerFunc{f})
		r.Add("GET", "/cbd/456", []HandlerFunc{f})
		r.Add("GET", "/cbd/346", []HandlerFunc{f})
	})
}

func TestTreeRouteAdd2(t *testing.T) {
	Convey("add param route", t, func() {
		r.Add("GET", "/a/:id/id", []HandlerFunc{f})
		r.Add("GET", "/a/:id/name", []HandlerFunc{f})
		r.Add("GET", "/a", []HandlerFunc{f})
		r.Add("GET", "/a/:id/", []HandlerFunc{f})
		r.Add("GET", "/a/", []HandlerFunc{f})
		r.Add("GET", "/a/*/xxx", []HandlerFunc{f})
		r.Add("GET", "/p/:project/file/:fileName", []HandlerFunc{f})
		r.Add("GET", "/cbd/:id", []HandlerFunc{f})

		defer func() {
			e := recover()
			So(e, ShouldNotBeNil)
		}()
		r.Add("GET", "/p/:/a", []HandlerFunc{f})
	})
}

func TestTreeRouteAdd3(t *testing.T) {
	Convey("add param route with two different param", t, func() {
		defer func() {
			e := recover()
			So(e, ShouldNotBeNil)
		}()
		r.Add("GET", "/a/:id", []HandlerFunc{f})
		r.Add("GET", "/a/:name", []HandlerFunc{f})
	})
}

func TestTreeRouteAdd4(t *testing.T) {
	Convey("add route by group", t, func() {
		n.Group("/user", func() {
			n.Get("/info", f)
			n.Get("/info2", f)
			n.Group("/group", func() {
				n.Get("/info", f)
				n.Get("/info2", f)
			})
		})
		n.Group("/user", func() {
			n.Get("/", f)
			n.Get("/pass", f)
			n.Get("/pass2", f)
		}, f)
	})
}

func TestTreeRouteAdd5(t *testing.T) {
	Convey("add route then set name, URLFor", t, func() {
		n.Get("/article/:id/show", f).Name("articleShow")
		n.Get("/article/:id/detail", f).Name("")
		url := n.URLFor("articleShow", 123)
		So(url, ShouldEqual, "/article/123/show")
		url = n.URLFor("", nil)
		So(url, ShouldEqual, "")
		url = n.URLFor("not exits", "no")
		So(url, ShouldEqual, "")
		ru, name := r.Match("GET", "/article/123/show", c)
		So(ru, ShouldNotBeNil)
		So(name, ShouldEqual, "articleShow")
	})
}

func TestTreeRouteAdd6(t *testing.T) {
	Convey("add route with not support method", t, func() {
		defer func() {
			e := recover()
			So(e, ShouldNotBeNil)
		}()
		r.Add("TRACE", "/", []HandlerFunc{f})
	})
}

func TestTreeRouteAdd7(t *testing.T) {
	Convey("add route with empty pattern", t, func() {
		defer func() {
			e := recover()
			So(e, ShouldNotBeNil)
		}()
		r.Add("GET", "", []HandlerFunc{f})
	})
}

func TestTreeRouteAdd8(t *testing.T) {
	Convey("add route with pattern that not begin with /", t, func() {
		defer func() {
			e := recover()
			So(e, ShouldNotBeNil)
		}()
		r.Add("GET", "abc", []HandlerFunc{f})
	})
}

func TestTreeRouteAdd9(t *testing.T) {
	Convey("other route method", t, func() {
		b2 := New()
		Convey("set auto head route", func() {
			b2.SetAutoHead(true)
			b2.Get("/head", func(c *Context) {
				So(c.Req.Method, ShouldEqual, "HEAD")
			})
			req, _ := http.NewRequest("HEAD", "/head", nil)
			w := httptest.NewRecorder()
			b2.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusOK)
		})
		Convey("set auto training slash", func() {
			b2.SetAutoTrailingSlash(true)
			b2.Get("/slash", func(c *Context) {})
			b2.Group("/slash2", func() {
				b2.Get("/", func(c *Context) {})
				b2.Get("/exist", func(c *Context) {})
			})
			req, _ := http.NewRequest("GET", "/slash", nil)
			w := httptest.NewRecorder()
			b2.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusOK)
			req, _ = http.NewRequest("GET", "/slash/", nil)
			w = httptest.NewRecorder()
			b2.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusOK)
			req, _ = http.NewRequest("GET", "/slash2", nil)
			w = httptest.NewRecorder()
			b2.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusOK)
			req, _ = http.NewRequest("GET", "/slash2/", nil)
			w = httptest.NewRecorder()
			b2.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusOK)
			req, _ = http.NewRequest("GET", "/slash2/exist/", nil)
			w = httptest.NewRecorder()
			b2.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusOK)
		})
		Convey("set multi method", func() {
			b2.Route("/mul1", "*", func(c *Context) {
				c.String(200, "mul")
			})
			b2.Route("/mul2", "GET,HEAD,POST", func(c *Context) {
				c.String(200, "mul")
			})
			req, _ := http.NewRequest("HEAD", "/mul2", nil)
			w := httptest.NewRecorder()
			b2.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusOK)

			req, _ = http.NewRequest("GET", "/mul2", nil)
			w = httptest.NewRecorder()
			b2.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusOK)
			req, _ = http.NewRequest("POST", "/mul2", nil)
			w = httptest.NewRecorder()
			b2.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusOK)
		})
		Convey("methods", func() {
			b2.Get("/methods", f)
			b2.Patch("/methods", f)
			b2.Post("/methods", f)
			b2.Put("/methods", f)
			b2.Delete("/methods", f)
			b2.Options("/methods", f)
			b2.Head("/methods", f)
			b2.Any("/any", f)
			b2.SetNotFound(func(c *Context) {
				c.String(404, "nice not found")
			})
		})
	})
}

func TestTreeRouteMatch1(t *testing.T) {
	Convey("match route", t, func() {

		ru, _ := r.Match("GET", "/", c)
		So(ru, ShouldNotBeNil)

		ru, _ = r.Match("GET", "/abc/1234", c)
		So(ru, ShouldBeNil)

		ru, _ = r.Match("GET", "xxx", c)
		So(ru, ShouldBeNil)

		ru, _ = r.Match("GET", "/a/123/id", c)
		So(ru, ShouldNotBeNil)

		ru, _ = r.Match("GET", "/p/yst/file/a.jpg", c)
		So(ru, ShouldNotBeNil)

		ru, _ = r.Match("GET", "/user/info", c)
		So(ru, ShouldNotBeNil)

		ru, _ = r.Match("GET", "/user/pass", c)
		So(ru, ShouldNotBeNil)

		ru, _ = r.Match("GET", "/user/pass32", c)
		So(ru, ShouldBeNil)

		ru, _ = r.Match("GET", "/user/xxx", c)
		So(ru, ShouldBeNil)

		ru, _ = r.Match("GET", "/xxxx", c)
		So(ru, ShouldBeNil)

		n.Get("/notifications/threads/:id", f)
		n.Get("/notifications/threads/:id/subscription", f)
		n.Get("/notifications/threads/:id/subc", f)
		n.Put("/notifications/threads/:id/subscription", f)
		n.Delete("/notifications/threads/:id/subscription", f)
		ru, _ = r.Match("GET", "/notifications/threads/:id", c)
		So(ru, ShouldNotBeNil)
		ru, _ = r.Match("GET", "/notifications/threads/:id/sub", c)
		So(ru, ShouldBeNil)
	})
}

func TestTreeRoutePrint1(t *testing.T) {
	Convey("print route table", t, func() {
		r.(*Tree).print("", nil)
	})
}

func TestTreeRoutePrint2(t *testing.T) {
	Convey("print routes", t, func() {
		fmt.Println("")
		for method, routes := range r.Routes() {
			fmt.Println("Method: ", method)
			for i, route := range routes {
				fmt.Printf(" %3d %s\n", i, route)
			}
		}
	})
}

func TestTreeRoutePrint3(t *testing.T) {
	Convey("print named routes", t, func() {
		fmt.Println("")
		for name, route := range r.NamedRoutes() {
			fmt.Printf("%20s \t %s\n", name, route)
		}
	})
}
