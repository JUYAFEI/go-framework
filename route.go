package go_framework

import (
	"fmt"
	"log"
	"net/http"
)

type HandlerFunc func(http.ResponseWriter, *http.Request)

type Router struct {
	groups []*RouterGroup
}

func (r *Router) Group(name string) *RouterGroup {
	g := &RouterGroup{
		groupName:        name,
		handlerMap:       make(map[string]HandlerFunc),
		handlerMethodMap: make(map[string][]string),
	}
	r.groups = append(r.groups, g)
	return g
}

type RouterGroup struct {
	groupName        string                 // group name
	handlerMap       map[string]HandlerFunc // handler map
	handlerMethodMap map[string][]string    // handler method map
}

func (r *RouterGroup) Add(name string, handlerFunc HandlerFunc) {
	r.handlerMap[name] = handlerFunc
}

type Engine struct {
	*Router
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.Method)
}

func New() *Engine {
	return &Engine{
		&Router{},
	}
}

func (e *Engine) Run() {
	//http.Handle("/", e)
	groups := e.Router.groups
	for _, g := range groups {
		for name, handle := range g.handlerMap {
			http.HandleFunc("/"+g.groupName+"/"+name, handle)
		}
	}
	err := http.ListenAndServe(":8111", nil)
	if err != nil {
		log.Fatal(err)
	}
}
