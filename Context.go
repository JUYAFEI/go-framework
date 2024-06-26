package go_framework

import (
	"encoding/json"
	"errors"
	golog "github.com/JUYAFEI/go-framework/log"
	"github.com/JUYAFEI/go-framework/render"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

const DefaultMemory = 32 << 20

type Context struct {
	W          http.ResponseWriter
	R          *http.Request
	Engine     *Engine
	queryCache url.Values
	formCache  url.Values
	StatusCode int
	Logger     *golog.Logger
	sameSite   http.SameSite
	mu         sync.RWMutex
	Keys       map[string]any
}

func (c *Context) SetSameSite(s http.SameSite) {
	c.sameSite = s
}
func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}
	c.Keys[key] = value
	c.mu.Unlock()
}

func (c *Context) Get(key string) (value any, ok bool) {
	c.mu.RLock()
	value, ok = c.Keys[key]
	c.mu.RUnlock()
	return
}

func (c *Context) QueryMap(key string) (dicts map[string]string) {
	dicts, _ = c.GetQueryMap(key)
	return
}

func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	c.initQueryCache()
	return c.get(c.queryCache, key)
}

func (c *Context) get(m map[string][]string, key string) (map[string]string, bool) {
	//user[id]=1&user[name]=张三
	dicts := make(map[string]string)
	exist := false
	for k, value := range m {
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				exist = true
				dicts[k[i+1:][:j]] = value[0]
			}
		}
	}
	return dicts, exist
}

func (c *Context) DealJson(data any) error {
	body := c.R.Body
	if c.R == nil || body == nil {
		return errors.New("invalid request")
	}
	decoder := json.NewDecoder(body)
	return decoder.Decode(data)
}

func (c *Context) initQueryCache() {
	if c.R != nil {
		c.queryCache = c.R.URL.Query()
	} else {
		c.queryCache = url.Values{}
	}
}

func (c *Context) DefaultQuery(key, defaultValue string) string {
	array, ok := c.GetQueryArray(key)
	if !ok {
		return defaultValue
	}
	return array[0]
}

func (c *Context) GetQueryArray(key string) (values []string, ok bool) {
	c.initQueryCache()
	values, ok = c.queryCache[key]
	return
}

func (c *Context) GetQuery(key string) string {
	c.initQueryCache()
	return c.queryCache.Get(key)
}

func (c *Context) QueryArray(key string) (values []string) {
	c.initQueryCache()
	values, _ = c.queryCache[key]
	return
}

func (c *Context) initPostFormCache() {
	if c.formCache == nil {
		c.formCache = make(url.Values)
		req := c.R
		if err := req.ParseMultipartForm(DefaultMemory); err != nil {
			if !errors.Is(err, http.ErrNotMultipart) {
				log.Println(err)
			}
		}
		c.formCache = c.R.PostForm
	}
}

func (c *Context) GetPostForm(key string) (string, bool) {
	if values, ok := c.GetPostFormArray(key); ok {
		return values[0], ok
	}
	return "", false
}

func (c *Context) PostFormArray(key string) (values []string) {
	values, _ = c.GetPostFormArray(key)
	return
}

func (c *Context) GetPostFormArray(key string) (values []string, ok bool) {
	c.initPostFormCache()
	values, ok = c.formCache[key]
	return
}

func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	c.initPostFormCache()
	return c.get(c.formCache, key)
}

func (c *Context) PostFormMap(key string) (dicts map[string]string) {
	dicts, _ = c.GetPostFormMap(key)
	return
}

func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	req := c.R
	if err := req.ParseMultipartForm(DefaultMemory); err != nil {
		return nil, err
	}
	file, header, err := req.FormFile(name)
	if err != nil {
		return nil, err
	}
	err = file.Close()
	if err != nil {
		return nil, err
	}
	return header, nil
}

func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.R.ParseMultipartForm(DefaultMemory)
	return c.R.MultipartForm, err
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
	c.StatusCode = statusCode
	if statusCode != http.StatusOK {
		c.W.WriteHeader(statusCode)
	}
	return err
}

func (c *Context) Fail(code int, msg string) {
	c.String(code, msg)
}

func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.W, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		SameSite: c.sameSite,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}
