package render

import (
	"html/template"
	"net/http"
)

type HTMLRender struct {
	Template *template.Template
}

type Render interface {
	Render(w http.ResponseWriter) error
	WriteContentType(w http.ResponseWriter)
}
