package render

import (
	"html/template"
	"net/http"
)

type HTMLData any

type HTML struct {
	Template   *template.Template
	Name       string
	Data       HTMLData
	IsTemplate bool
}

var htmlContentType = []string{"text/html; charset=utf-8"}

func (r HTML) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	if !r.IsTemplate {
		_, err := w.Write([]byte(r.Data.(string)))
		return err
	}
	err := r.Template.ExecuteTemplate(w, r.Name, r.Data)
	return err
}

func (r HTML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "text/html; charset=utf-8")
}
