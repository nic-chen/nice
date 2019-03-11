package nice

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

const (
	// DEV mode
	DEV = "development"
	// PROD mode
	PROD = "production"
	// TEST mode
	TEST = "test"
)

// Env default application runtime environment
var Env string

// Nice provlider an application
type Nice struct {
	debug           bool
	name            string
	di              DIer
	router          Router
	pool            sync.Pool
	errorHandler    ErrorHandleFunc
	notFoundHandler HandlerFunc
	middleware      []HandlerFunc
}

// Middleware middleware handler
type Middleware interface{}

// HandlerFunc context handler func
type HandlerFunc func(*Context)

// ErrorHandleFunc HTTP error handleFunc
type ErrorHandleFunc func(error, *Context)

// appInstances storage application instances
var appInstances map[string]*Nice

// default application name
const default_app_name = "nice"

// New create a nice application without any config.
func New() *Nice {
	n := new(Nice)
	n.middleware = make([]HandlerFunc, 0)
	n.pool = sync.Pool{
		New: func() interface{} {
			return NewContext(nil, nil, n)
		},
	}
	if Env != PROD {
		n.debug = true
	}
	n.SetDIer(NewDI())
	n.SetDI("router", NewTree(n))
	n.SetDI("logger", log.New(os.Stderr, "[Nice] ", log.LstdFlags))
	n.SetDI("render", newRender())
	// n.SetDI("db", NewMysql())
	// n.SetDI("cache", NewRedis())
	n.SetNotFound(n.DefaultNotFoundHandler)
	return n
}

// Instance register or returns named application
func Instance(name string) *Nice {
	if name == "" {
		name = default_app_name
	}
	if appInstances[name] == nil {
		appInstances[name] = New()
		appInstances[name].name = name
	}
	return appInstances[name]
}

// Default initial a default app then returns
func Default() *Nice {
	return Instance(default_app_name)
}

// Server returns the internal *http.Server.
func (n *Nice) Server(addr string) *http.Server {
	s := &http.Server{Addr: addr}
	return s
}

// Run runs a server.
func (n *Nice) Run(addr string) {
	n.run(n.Server(addr))
}

// RunTLS runs a server with TLS configuration.
func (n *Nice) RunTLS(addr, certfile, keyfile string) {
	n.run(n.Server(addr), certfile, keyfile)
}

// RunServer runs a custom server.
func (n *Nice) RunServer(s *http.Server) {
	n.run(s)
}

// RunTLSServer runs a custom server with TLS configuration.
func (n *Nice) RunTLSServer(s *http.Server, crtFile, keyFile string) {
	n.run(s, crtFile, keyFile)
}

func (n *Nice) run(s *http.Server, files ...string) {
	s.Handler = n
	n.Logger().Printf("Run mode: %s", Env)
	if len(files) == 0 {
		n.Logger().Printf("Listen %s", s.Addr)
		n.Logger().Fatal(s.ListenAndServe())
	} else if len(files) == 2 {
		n.Logger().Printf("Listen %s with TLS", s.Addr)
		n.Logger().Fatal(s.ListenAndServeTLS(files[0], files[1]))
	} else {
		panic("invalid TLS configuration")
	}
}

func (n *Nice) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := n.pool.Get().(*Context)
	c.Reset(w, r)

	// build handler chain
	h, name := n.Router().Match(r.Method, r.URL.Path, c)
	c.routeName = name

	// notFound
	if h == nil {
		c.handlers = append(c.handlers, n.notFoundHandler)
	} else {
		c.handlers = append(c.handlers, h...)
	}

	c.Next()

	n.pool.Put(c)
}

// SetDIer set nice di
func (n *Nice) SetDIer(v DIer) {
	n.di = v
}

// SetDebug set nice debug
func (n *Nice) SetDebug(v bool) {
	n.debug = v
}

// Debug returns nice debug state
func (n *Nice) Debug() bool {
	return n.debug
}

// Logger return nice logger
func (n *Nice) Logger() Logger {
	return n.GetDI("logger").(Logger)
}

//db
func (n *Nice) Db() Db {
	return n.GetDI("db").(Db)
}

func (n *Nice) Cache() Cache {
	return n.GetDI("cache").(Cache)
}

// Render return nice render
func (n *Nice) Render() Renderer {
	return n.GetDI("render").(Renderer)
}

// Router return nice router
func (n *Nice) Router() Router {
	if n.router == nil {
		n.router = n.GetDI("router").(Router)
	}
	return n.router
}

// Use registers a middleware
func (n *Nice) Use(m ...Middleware) {
	for i := range m {
		if m[i] != nil {
			n.middleware = append(n.middleware, wrapMiddleware(m[i]))
		}
	}
}

// SetDI registers a dependency injection
func (n *Nice) SetDI(name string, h interface{}) {
	switch name {
	case "logger":
		if _, ok := h.(Logger); !ok {
			panic("DI logger must be implement interface nice.Logger")
		}
	case "render":
		if _, ok := h.(Renderer); !ok {
			panic("DI render must be implement interface nice.Renderer")
		}
	case "router":
		if _, ok := h.(Router); !ok {
			panic("DI router must be implement interface nice.Router")
		}
	}
	n.di.Set(name, h)
}

// GetDI fetch a registered dependency injection
func (n *Nice) GetDI(name string) interface{} {
	return n.di.Get(name)
}

// Static set static file route
// h used for set Expries ...
func (n *Nice) Static(prefix string, dir string, index bool, h HandlerFunc) {
	if prefix == "" {
		panic("nice.Static prefix can not be empty")
	}
	if dir == "" {
		panic("nice.Static dir can not be empty")
	}
	n.Get(prefix+"*", newStatic(prefix, dir, index, h))
}

// StaticFile shortcut for serve file
func (n *Nice) StaticFile(pattern string, path string) RouteNode {
	return n.Get(pattern, func(c *Context) {
		if err := serveFile(path, c); err != nil {
			c.Error(err)
		}
	})
}

// SetAutoHead sets the value who determines whether add HEAD method automatically
// when GET method is added. Combo router will not be affected by this value.
func (n *Nice) SetAutoHead(v bool) {
	n.Router().SetAutoHead(v)
}

// SetAutoTrailingSlash optional trailing slash.
func (n *Nice) SetAutoTrailingSlash(v bool) {
	n.Router().SetAutoTrailingSlash(v)
}

// Route is a shortcut for same handlers but different HTTP methods.
//
// Example:
// 		nice.Route("/", "GET,POST", h)
func (n *Nice) Route(pattern, methods string, h ...HandlerFunc) RouteNode {
	var ru RouteNode
	var ms []string
	if methods == "*" {
		for m := range RouterMethods {
			ms = append(ms, m)
		}
	} else {
		ms = strings.Split(methods, ",")
	}
	for _, m := range ms {
		ru = n.Router().Add(strings.TrimSpace(m), pattern, h)
	}
	return ru
}

// Group registers a list of same prefix route
func (n *Nice) Group(pattern string, f func(), h ...HandlerFunc) {
	n.Router().GroupAdd(pattern, f, h)
}

// Any is a shortcut for n.Router().handle("*", pattern, handlers)
func (n *Nice) Any(pattern string, h ...HandlerFunc) RouteNode {
	var ru RouteNode
	for m := range RouterMethods {
		ru = n.Router().Add(m, pattern, h)
	}
	return ru
}

// Delete is a shortcut for n.Route(pattern, "DELETE", handlers)
func (n *Nice) Delete(pattern string, h ...HandlerFunc) RouteNode {
	return n.Router().Add("DELETE", pattern, h)
}

// Get is a shortcut for n.Route(pattern, "GET", handlers)
func (n *Nice) Get(pattern string, h ...HandlerFunc) RouteNode {
	return n.Router().Add("GET", pattern, h)
}

// Head is a shortcut forn.Route(pattern, "Head", handlers)
func (n *Nice) Head(pattern string, h ...HandlerFunc) RouteNode {
	return n.Router().Add("HEAD", pattern, h)
}

// Options is a shortcut for n.Route(pattern, "Options", handlers)
func (n *Nice) Options(pattern string, h ...HandlerFunc) RouteNode {
	return n.Router().Add("OPTIONS", pattern, h)
}

// Patch is a shortcut for n.Route(pattern, "PATCH", handlers)
func (n *Nice) Patch(pattern string, h ...HandlerFunc) RouteNode {
	return n.Router().Add("PATCH", pattern, h)
}

// Post is a shortcut for n.Route(pattern, "POST", handlers)
func (n *Nice) Post(pattern string, h ...HandlerFunc) RouteNode {
	return n.Router().Add("POST", pattern, h)
}

// Put is a shortcut for n.Route(pattern, "Put", handlers)
func (n *Nice) Put(pattern string, h ...HandlerFunc) RouteNode {
	return n.Router().Add("PUT", pattern, h)
}

// SetNotFound set not found route handler
func (n *Nice) SetNotFound(h HandlerFunc) {
	n.notFoundHandler = h
}

// NotFound execute not found handler
func (n *Nice) NotFound(c *Context) {
	if n.notFoundHandler != nil {
		n.notFoundHandler(c)
		return
	}
	http.NotFound(c.Resp, c.Req)
}

// SetError set error handler
func (n *Nice) SetError(h ErrorHandleFunc) {
	n.errorHandler = h
}

// Error execute internal error handler
func (n *Nice) Error(err error, c *Context) {
	if err == nil {
		err = errors.New("Internal Server Error")
	}
	if n.errorHandler != nil {
		n.errorHandler(err, c)
		return
	}
	code := http.StatusInternalServerError
	msg := http.StatusText(code)
	if n.debug {
		msg = err.Error()
	}
	n.Logger().Println(err)
	http.Error(c.Resp, msg, code)
}

// DefaultNotFoundHandler invokes the default HTTP error handler.
func (n *Nice) DefaultNotFoundHandler(c *Context) {
	code := http.StatusNotFound
	msg := http.StatusText(code)
	http.Error(c.Resp, msg, code)
}

// URLFor use named route return format url
func (n *Nice) URLFor(name string, args ...interface{}) string {
	return n.Router().URLFor(name, args...)
}

//加载配置 绝对路径
func (n *Nice) LoadConfig(yamlfile string) map[string]interface{} {
	conf := make(map[string]interface{})

	yamlFile, err := ioutil.ReadFile(yamlfile)
	if err != nil {
		log.Panicln("Read db config file failed.", err.Error())
	}

	log.Printf("c:%v", string(yamlFile))

	err = yaml.Unmarshal(yamlFile, conf)

	if err != nil {
		log.Panicln("Parse db config file failed.", err.Error())
	}

	log.Printf("conf:%v", conf)

	return conf
}

// wrapMiddleware wraps middleware.
func wrapMiddleware(m Middleware) HandlerFunc {
	switch m := m.(type) {
	case HandlerFunc:
		return m
	case func(*Context):
		return m
	case http.Handler, http.HandlerFunc:
		return WrapHandlerFunc(func(c *Context) {
			m.(http.Handler).ServeHTTP(c.Resp, c.Req)
		})
	case func(http.ResponseWriter, *http.Request):
		return WrapHandlerFunc(func(c *Context) {
			m(c.Resp, c.Req)
		})
	default:
		panic("unknown middleware")
	}
}

// WrapHandlerFunc wrap for context handler chain
func WrapHandlerFunc(h HandlerFunc) HandlerFunc {
	return func(c *Context) {
		h(c)
		c.Next()
	}
}

func init() {
	appInstances = make(map[string]*Nice)
	Env = os.Getenv("NICE_ENV")
	if Env == "" {
		Env = PROD
	}
}
