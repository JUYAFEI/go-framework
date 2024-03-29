package go_framework

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const Any = "Any"

type Context struct {
	W http.ResponseWriter
	R *http.Request
}

type HandlerFunc func(ctx *Context)

type Router struct {
	groups []*RouterGroup
}

func (r *Router) Group(name string) *RouterGroup {
	g := &RouterGroup{
		groupName:        name,
		handlerMap:       make(map[string]map[string]HandlerFunc),
		handlerMethodMap: make(map[string][]string),
		treeNode:         &Tree{Name: "/", Children: make([]*Tree, 0)},
	}
	r.groups = append(r.groups, g)
	return g
}

type RouterGroup struct {
	groupName        string                            // group name
	handlerMap       map[string]map[string]HandlerFunc // handler map
	handlerMethodMap map[string][]string               // handler method map
	treeNode         *Tree
}

func SubStringLast(str string, substr string) string {
	index := strings.Index(str, substr)
	if index == -1 {
		return ""
	}
	len := len(substr)
	return str[index+len:]
}

func (r *RouterGroup) handle(name string, method string, handlerFunc HandlerFunc) {
	_, ok := r.handlerMap[name]
	if !ok {
		r.handlerMap[name] = make(map[string]HandlerFunc)
	}
	r.handlerMap[name][method] = handlerFunc
	r.handlerMethodMap[method] = append(r.handlerMethodMap[method], name)
	r.treeNode.Put(name)
}

func (r *RouterGroup) Any(name string, handlerFunc HandlerFunc) {
	r.handle(name, Any, handlerFunc)
}

func (r *RouterGroup) Get(name string, handlerFunc HandlerFunc) {
	r.handle(name, http.MethodGet, handlerFunc)
}

func (r *RouterGroup) Post(name string, handlerFunc HandlerFunc) {
	r.handle(name, http.MethodPost, handlerFunc)
}

func (r *RouterGroup) Delete(name string, handlerFunc HandlerFunc) {
	r.handle(name, http.MethodDelete, handlerFunc)
}
func (r *RouterGroup) Put(name string, handlerFunc HandlerFunc) {
	r.handle(name, http.MethodPut, handlerFunc)
}
func (r *RouterGroup) Patch(name string, handlerFunc HandlerFunc) {
	r.handle(name, http.MethodPatch, handlerFunc)
}
func (r *RouterGroup) Options(name string, handlerFunc HandlerFunc) {
	r.handle(name, http.MethodOptions, handlerFunc)
}
func (r *RouterGroup) Head(name string, handlerFunc HandlerFunc) {
	r.handle(name, http.MethodHead, handlerFunc)
}

type Engine struct {
	*Router
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	groups := e.Router.groups
	for _, g := range groups {
		routerName := SubStringLast(req.RequestURI, "/"+g.groupName)
		node := g.treeNode.Get(routerName)
		if node != nil {
			ctx := &Context{
				W: w,
				R: req,
			}
			_, ok := g.handlerMap[node.RouterName][Any]
			if ok {
				g.handlerMap[node.RouterName][Any](ctx)
				return
			}
			method := req.Method
			handler, ok := g.handlerMap[node.RouterName][method]
			if ok {
				handler(ctx)
				return
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintln(w, req.RequestURI+method+" not allowed")
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "%s  not found \n", req.RequestURI)
}

func New() *Engine {
	return &Engine{
		&Router{},
	}
}

func (e *Engine) Run() {
	http.Handle("/", e)
	err := http.ListenAndServe(":8111", nil)
	if err != nil {
		log.Fatal(err)
	}
}
