package go_framework

import (
	"github.com/JUYAFEI/go-framework/render"
	"html/template"
	"log"
	"net/http"
	"net/url"
)

type Context struct {
	W      http.ResponseWriter
	R      *http.Request
	Engine *Engine
}

func (c *Context) HTML(status int, html string) {
	c.Render(status, render.HTML{IsTemplate: false, Data: html})
}

func (c *Context) HTMLTemplate(name string, data any) {
	c.Render(http.StatusOK, render.HTML{
		IsTemplate: true,
		Name:       name,
		Data:       data,
		Template:   c.Engine.HTMLRender.Template,
	})
}

func (c *Context) HTMLTemplateGlob(name string, funcMap template.FuncMap, pattern string, data any) {
	t := template.New(name)
	t.Funcs(funcMap)
	t, err := t.ParseGlob(pattern)
	if err != nil {
		log.Println(err)
		return
	}
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = t.Execute(c.W, data)
	if err != nil {
		log.Println(err)
	}
}

func (c *Context) Template(name string, data any) {
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	template := c.Engine.HTMLRender.Template
	err := template.ExecuteTemplate(c.W, name, data)
	if err != nil {
		log.Println(err)
	}
}

func (c *Context) JSON(code int, data any) error {
	return c.Render(code, render.JSON{Data: data})

}
func (c *Context) XML(status int, data any) error {
	return c.Render(status, render.XML{Data: data})
}

func (c *Context) File(filePath string) {
	http.ServeFile(c.W, c.R, filePath)
}

func (c *Context) FileAttachment(filepath, filename string) {
	if isASCII(filename) {
		c.W.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	} else {
		c.W.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	http.ServeFile(c.W, c.R, filepath)
}

func (c *Context) FileFromFS(filepath string, fs http.FileSystem) {
	defer func(old string) {
		c.R.URL.Path = old
	}(c.R.URL.Path)

	c.R.URL.Path = filepath

	http.FileServer(fs).ServeHTTP(c.W, c.R)
}

func (c *Context) Redirect(status int, location string) {
	c.Render(status, render.Redirect{
		Code:     status,
		Request:  c.R,
		Location: location,
	})
}

func (c *Context) String(status int, format string, values ...any) (err error) {
	err = c.Render(status, &render.String{
		Format: format,
		Data:   values,
	})
	return
}

func (c *Context) Render(statusCode int, r render.Render) error {
	err := r.Render(c.W)
	c.W.WriteHeader(statusCode)
	return err
}
