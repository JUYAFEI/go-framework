package go_framework

import (
	"fmt"
	"github.com/JUYAFEI/go-framework/render"
	"html/template"
	"log"
	"net/http"
	"sync"
)

const Any = "Any"

type HandlerFunc func(ctx *Context)

// MiddlewareFunc 定义中间件
type MiddlewareFunc func(handler HandlerFunc) HandlerFunc

type Router struct {
	groups []*RouterGroup
}

func (r *Router) Group(name string) *RouterGroup {
	g := &RouterGroup{
		groupName:          name,
		handlerMap:         make(map[string]map[string]HandlerFunc),
		middlewaresFuncMap: make(map[string]map[string][]MiddlewareFunc),
		handlerMethodMap:   make(map[string][]string),
		treeNode:           &Tree{Name: "/", Children: make([]*Tree, 0)},
	}
	r.groups = append(r.groups, g)
	return g
}

type RouterGroup struct {
	groupName          string                                 // group name
	handlerMap         map[string]map[string]HandlerFunc      // handler map
	middlewaresFuncMap map[string]map[string][]MiddlewareFunc // 中间件map
	handlerMethodMap   map[string][]string                    // handler method map
	treeNode           *Tree
	middlewares        []MiddlewareFunc // 前置中间件
}

func (r *RouterGroup) Use(middlewares ...MiddlewareFunc) {
	r.middlewares = append(r.middlewares, middlewares...)
}

func (r *RouterGroup) methodHandle(name string, method string, h HandlerFunc, ctx *Context) {
	// 通用中间件中间件
	if r.middlewares != nil {
		for _, middleware := range r.middlewares {
			h = middleware(h)
		}
	}
	// 路由中间件
	middlewareFuncs := r.middlewaresFuncMap[name][method]
	if middlewareFuncs != nil {
		for _, middlewareFunc := range middlewareFuncs {
			h = middlewareFunc(h)
		}
	}
	h(ctx)
}

func (r *RouterGroup) handle(name string, method string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	_, ok := r.handlerMap[name]
	if !ok {
		r.handlerMap[name] = make(map[string]HandlerFunc)
		r.middlewaresFuncMap[name] = make(map[string][]MiddlewareFunc)
	}
	r.handlerMap[name][method] = handlerFunc
	r.handlerMethodMap[method] = append(r.handlerMethodMap[method], name)
	r.middlewaresFuncMap[name][method] = append(r.middlewaresFuncMap[name][method], middlewareFunc...)
	r.treeNode.Put(name)
}

func (r *RouterGroup) Any(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, Any, handlerFunc, middlewareFunc...)
}

func (r *RouterGroup) Get(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodGet, handlerFunc, middlewareFunc...)
}

func (r *RouterGroup) Post(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodPost, handlerFunc, middlewareFunc...)
}

func (r *RouterGroup) Delete(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodDelete, handlerFunc, middlewareFunc...)
}
func (r *RouterGroup) Put(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodPut, handlerFunc, middlewareFunc...)
}
func (r *RouterGroup) Patch(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodPatch, handlerFunc, middlewareFunc...)
}
func (r *RouterGroup) Options(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodOptions, handlerFunc, middlewareFunc...)
}
func (r *RouterGroup) Head(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodHead, handlerFunc, middlewareFunc...)
}

type Engine struct {
	*Router
	funcMap    template.FuncMap
	HTMLRender render.HTMLRender
	pool       sync.Pool
}

func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

func (e *Engine) SetHtmlTemplate(t *template.Template) {
	e.HTMLRender = render.HTMLRender{Template: t}
}

func (e *Engine) LoadTemplateGlob(pattern string) {
	t := template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
	e.SetHtmlTemplate(t)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := e.pool.Get().(*Context)
	ctx.W = w
	ctx.R = req
	e.httpRequestHandler(ctx)
}

func New() *Engine {

	engine := &Engine{
		Router:     &Router{},
		funcMap:    nil,
		HTMLRender: render.HTMLRender{},
	}
	engine.pool.New = func() any {
		return engine.allocateContext()
	}
	return engine
}

func (e *Engine) allocateContext() *Context {
	return &Context{Engine: e}
}

func (e *Engine) httpRequestHandler(ctx *Context) {
	groups := e.Router.groups
	for _, g := range groups {
		routerName := SubStringLast(ctx.R.URL.Path, "/"+g.groupName)
		node := g.treeNode.Get(routerName)
		if node != nil {
			anyHandler, ok := g.handlerMap[node.RouterName][Any]
			if ok {
				g.methodHandle(node.RouterName, Any, anyHandler, ctx)
				return
			}
			method := ctx.R.Method
			handler, ok := g.handlerMap[node.RouterName][method]
			if ok {
				g.methodHandle(node.RouterName, method, handler, ctx)
				return
			}
			ctx.W.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintln(ctx.W, ctx.R.RequestURI+method+" not allowed")
			return
		}
	}
	ctx.W.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(ctx.W, "%s  not found \n", ctx.R.RequestURI)
}

func (e *Engine) Run() {
	http.Handle("/", e)
	err := http.ListenAndServe(":8111", nil)
	if err != nil {
		log.Fatal(err)
	}
}
