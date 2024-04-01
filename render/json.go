package render

import (
	"encoding/json"
	"net/http"
)

type JSON struct {
	Data any
}

func (r JSON) Render(w http.ResponseWriter) error {
	return WriteJSON(w, r.Data)
}
func (r JSON) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "application/json; charset=utf-8")
}

func WriteJSON(w http.ResponseWriter, obj any) error {
	writeContentType(w, "application/json; charset=utf-8")
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonBytes)
	return err
}
